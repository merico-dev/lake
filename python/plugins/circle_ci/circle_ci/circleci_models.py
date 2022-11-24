from pydevlake.core.models import *


@dataclass
class Connection(BaseConnection):
    endpoint: str = ""
    rateLimitPerHour: int = 0
    token: Field[str] | str = Field(value="", tags='encrypt:"yes"')
    proxy: str = ""

    def get_doc_schema(self):
        return ConnectionSchema()


class ConnectionSchema(BaseConnectionSchema):
    endpoint = fields.Str()
    rateLimitPerHour = fields.Number()
    token = fields.Str()
    proxy = fields.Str()


class ConnectionListSchema(Schema):
    connections = fields.List(fields.Nested(ConnectionSchema))
