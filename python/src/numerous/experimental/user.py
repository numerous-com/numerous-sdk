from numerous.experimental.framework_detection import FrameworkDetector
from numerous.collection import collection

class User:
    def __init__(self):
        self.framework_detection = FrameworkDetector()

    @property
    def id(self):
        cookies = self.framework_detection.get_cookies()
        return cookies.get('numerous_user_id', "local-user")

    def collection(self, collection_key: str):
        return collection("users").collection(self.id).collection(collection_key)
