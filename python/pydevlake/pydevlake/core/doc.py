from typing import Callable, Iterable, Type

from .doc_models import *
from .swagger import docgen


def __resolve_schema_names__(sch) -> tuple[str, str]:
    sch_name: str = sch.__class__.__name__.replace(".", "/")  # or use fully qualified name?
    # marshmallow framework workaround
    sch_name = sch_name.removesuffix("Schema") if sch_name.endswith("Schema") else sch_name
    sch_list_name: str = (sch_name + "List").replace(".", "/")
    return sch_name, sch_list_name


def normalize_doc_types(*types: Iterable[Type[DocSchema]]):
    normalized = []
    for t in types:
        try:
            from marshmallow import fields
            if isinstance(t, Iterable):
                sch = t[0]().get_doc_schema()
                sch_name, sch_list_name = __resolve_schema_names__(sch)
                from .swagger.docgen import spec
                # this is a hack, but I don't know a better way
                # TODO should technically check if the 'singular' type is registered too - it's a rare edge case to hit
                spec.components.schemas.update({
                    sch_list_name: {
                        "type": "array",
                        "items": {
                            "$ref": '#/components/schemas/{}'.format(sch_name)
                        }
                    }
                })
                normalized.append(sch_list_name)
            else:
                # from .ipc import start_debugger
                # start_debugger()
                sch = t().get_doc_schema()
                type_name = sch.__class__.__name__
                normalized.append(type_name)
        except AttributeError:
            raise
    return normalized


def api_doc(path: str, doc_str: str, *types):
    def fn(func: Callable):
        def wrapper(*args, **kwargs):
            return func(*args, **kwargs)

        doc_gen = func.__annotations__.get("docgen")
        if doc_gen is not True:
            if doc_str != "":
                resolved = normalize_doc_types(*types)
                func.__doc__ = doc_str.format(*resolved)
                docgen.generate_doc("/plugins/{}".format(path), func)
            func.__annotations__.__setitem__("docgen", True)
        return wrapper

    return fn
