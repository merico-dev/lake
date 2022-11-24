from apispec import *
from apispec.ext.marshmallow import MarshmallowPlugin
from apispec.yaml_utils import load_operations_from_docstring

spec = APISpec(
    title="Swagger Docs", version="1.0.0", openapi_version="3.0.2", plugins=[MarshmallowPlugin()]
)


def generate_doc(path: str = "", func=None):
    spec.path(path=path, operations=load_operations_from_docstring(func.__doc__))
