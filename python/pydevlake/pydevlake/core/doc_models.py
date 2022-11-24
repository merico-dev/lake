import abc

from marshmallow import fields, Schema


class DocSchema(metaclass=abc.ABCMeta):
    @abc.abstractmethod
    def get_doc_schema(self) -> Schema:
        pass


class BaseConnectionSchema(Schema):
    Name = fields.Str()
    ID = fields.Number()
