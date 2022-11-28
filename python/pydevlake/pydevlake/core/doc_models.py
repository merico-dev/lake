import abc
from dataclasses import dataclass

from marshmallow import fields, Schema


class DocSchema(metaclass=abc.ABCMeta):
    @abc.abstractmethod
    def get_doc_schema(self) -> Schema:
        pass


class BaseConnectionSchema(Schema):
    Name = fields.Str()
    ID = fields.Number()


class ApiDoc:

    def __init__(self, path: str, doc: str, *types):
        self.path = path
        self.doc = doc
        self.types = types
