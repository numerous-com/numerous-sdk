# Generated by ariadne-codegen
# Source: ../shared/schema.gql

from enum import Enum


class AuthRole(str, Enum):
    AUTHENTICATED = "AUTHENTICATED"
    ADMIN = "ADMIN"
    USER = "USER"


class Role(str, Enum):
    ADMIN = "ADMIN"
    USER = "USER"


class AppDeploymentStatus(str, Enum):
    PENDING = "PENDING"
    RUNNING = "RUNNING"
    ERROR = "ERROR"
    STOPPED = "STOPPED"
    UNKNOWN = "UNKNOWN"


class SubscriptionOfferStatus(str, Enum):
    ACCEPTED = "ACCEPTED"
    WITHDRAWN = "WITHDRAWN"
    REJECTED = "REJECTED"
    PENDING = "PENDING"


class AppSubscriptionStatus(str, Enum):
    ACTIVE = "ACTIVE"
    WITHDRAWN = "WITHDRAWN"
    EXPIRED = "EXPIRED"
    CANCELED = "CANCELED"


class ToolHashType(str, Enum):
    public = "public"
    shared = "shared"
    private = "private"


class PaymentAccountStatus(str, Enum):
    RESTRICTED = "RESTRICTED"
    VERIFIED = "VERIFIED"
    UNKNOWN = "UNKNOWN"