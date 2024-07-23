import asyncio
from numerous.generated.graphql.fragments import CollectionKey

from numerous.generated.graphql.client import Client


class NumerousClient:

    def __init__(self, client: Client):
        self.client = client
        self.organization_id = "" 
        
    def get_collection_key(self,collection_key: str) ->CollectionKey:
        cc =  asyncio.run(self.client.collection_create(self.organization_id,collection_key,None))
        return cc.collection_create
    
    def get_collection_key_with_parent(self,collection_key: str,parent_collection_key:str) ->CollectionKey:
        cc =   asyncio.run(self.client.collection_create(self.organization_id,collection_key,parent_collection_key))
        return cc.collection_create



def open_client(access_token: str) -> NumerousClient:
    return NumerousClient(Client(url=""))
