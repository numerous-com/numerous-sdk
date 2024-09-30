import os
import json
from typing import Optional, List, Tuple, Dict
from numerous.generated.graphql.fragments import CollectionReference, CollectionDocumentReference
from numerous.generated.graphql.input_types import TagInput
from numerous.jsonbase64 import dict_to_base64, base64_to_dict

class LocalClient:
    def __init__(self, base_path: str = '.collection'):
        self.base_path = base_path
        os.makedirs(self.base_path, exist_ok=True)

    def get_collection_reference(
        self, collection_key: str, parent_collection_id: Optional[str] = None
    ) -> Optional[CollectionReference]:
        if parent_collection_id:
            collection_id = os.path.join(parent_collection_id, collection_key)
        else:
            collection_id = collection_key

        print('coll: ', collection_id)
        path = os.path.join(self.base_path, collection_id)
        os.makedirs(path, exist_ok=True)
        return CollectionReference(id=collection_id, key=collection_key)

    def get_collection_document(
        self, collection_id: str, document_key: str
    ) -> Optional[CollectionDocumentReference]:
        print("col id!: ",collection_id)
        path = os.path.join(self.base_path, collection_id, f"{document_key}.json")
        if not os.path.exists(path):
            # Create an empty document if it doesn't exist
            empty_data = {}
            os.makedirs(os.path.dirname(path), exist_ok=True)
            with open(path, 'w') as f:
                json.dump({'data': empty_data, 'tags': []}, f)
        
        with open(path, 'r') as f:
            file_content = json.load(f)
        
        return CollectionDocumentReference(
            id=os.path.join(collection_id, document_key),
            key=document_key,
            data=dict_to_base64(file_content['data']),
            tags=file_content.get('tags', [])
        )

    def set_collection_document(
        self, collection_id: str, document_key: str, document_data: str
    ) -> Optional[CollectionDocumentReference]:
        path = os.path.join(self.base_path, collection_id, f"{document_key}.json")
        data = {'data': base64_to_dict(document_data), 'tags': []}
        os.makedirs(os.path.dirname(path), exist_ok=True)
        with open(path, 'w') as f:
            json.dump(data, f)
        return CollectionDocumentReference(
            id=os.path.join(collection_id, document_key),
            key=document_key,
            data=document_data,
            tags=[]
        )

    def delete_collection_document(
        self, document_id: str
    ) -> Optional[CollectionDocumentReference]:
        if os.path.exists(document_id):
            with open(document_id, 'r') as f:
                data = json.load(f)
            os.remove(document_id)
            return CollectionDocumentReference(
                id=document_id,
                key=os.path.splitext(os.path.basename(document_id))[0],
                data=json.dumps(data.get('data', {})),
                tags=data.get('tags', [])
            )
        return None

    def add_collection_document_tag(
        self, document_id: str, tag: TagInput
    ) -> Optional[CollectionDocumentReference]:
        if os.path.exists(document_id):
            with open(document_id, 'r') as f:
                data = json.load(f)
            data['tags'].append(tag)
            with open(document_id, 'w') as f:
                json.dump(data, f)
            return CollectionDocumentReference(
                id=document_id,
                key=os.path.splitext(os.path.basename(document_id))[0],
                data=json.dumps(data.get('data', {})),
                tags=data['tags']
            )
        return None

    def delete_collection_document_tag(
        self, document_id: str, tag_key: str
    ) -> Optional[CollectionDocumentReference]:
        if os.path.exists(document_id):
            with open(document_id, 'r') as f:
                data = json.load(f)
            data['tags'] = [tag for tag in data['tags'] if tag['key'] != tag_key]
            with open(document_id, 'w') as f:
                json.dump(data, f)
            return CollectionDocumentReference(
                id=document_id,
                key=os.path.splitext(os.path.basename(document_id))[0],
                data=json.dumps(data.get('data', {})),
                tags=data['tags']
            )
        return None

    def get_collection_documents(
        self, collection_key: str, end_cursor: str, tag_input: Optional[TagInput]
    ) -> Tuple[Optional[List[Optional[CollectionDocumentReference]]], bool, str]:
        path = os.path.join(self.base_path, collection_key)
        if not os.path.exists(path):
            return [], False, ""
        
        documents = []
        for filename in os.listdir(path):
            if filename.endswith('.json'):
                doc_path = os.path.join(path, filename)
                with open(doc_path, 'r') as f:
                    data = json.load(f)
                
                # Check if the document matches the tag_input
                if tag_input:
                    matching_tag = next((tag for tag in data.get('tags', []) 
                                         if tag['key'] == tag_input.key and tag['value'] == tag_input.value), None)
                    if not matching_tag:
                        continue

                documents.append(CollectionDocumentReference(
                    id=doc_path,
                    key=os.path.splitext(filename)[0],
                    data=dict_to_base64(data.get('data', {})),
                    tags=data.get('tags', [])
                ))
        
        return documents, False, ""

    def get_collection_collections(
        self, collection_key: str, end_cursor: str
    ) -> Tuple[Optional[List[Optional[CollectionReference]]], bool, str]:
        path = os.path.join(self.base_path, collection_key)
        if not os.path.exists(path):
            return [], False, ""
        
        collections = []
        for item in os.listdir(path):
            item_path = os.path.join(path, item)
            if os.path.isdir(item_path):
                collections.append(CollectionReference(
                    id=item_path,
                    key=item
                ))
        
        return collections, False, ""

    def get_documents_by_tag(self, tag_input: TagInput) -> List[CollectionDocumentReference]:
        """
        Recursively search for documents with the specified tag across all collections.
        """
        documents = []
        self._recursive_tag_search(self.base_path, tag_input, documents)
        return documents

    def _recursive_tag_search(self, path: str, tag_input: TagInput, documents: List[CollectionDocumentReference]):
        for item in os.listdir(path):
            item_path = os.path.join(path, item)
            if os.path.isdir(item_path):
                # Recursively search subdirectories
                self._recursive_tag_search(item_path, tag_input, documents)
            elif item.endswith('.json'):
                # Check if the document has the specified tag
                with open(item_path, 'r') as f:
                    data = json.load(f)
                
                matching_tag = next((tag for tag in data.get('tags', []) 
                                     if tag['key'] == tag_input.key and tag['value'] == tag_input.value), None)
                if matching_tag:
                    documents.append(CollectionDocumentReference(
                        id=os.path.relpath(item_path, self.base_path),
                        key=os.path.splitext(item)[0],
                        data=dict_to_base64(data.get('data', {})),
                        tags=data.get('tags', [])
                    ))