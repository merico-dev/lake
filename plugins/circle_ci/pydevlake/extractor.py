from typing import Type
from pydevlake import ToolModel


def autoextract(json: dict, model_cls: Type[ToolModel]) -> ToolModel:
    annotations = dict(model_cls.__annotations__)
    for key, value in json.items():
        if key in annotations:
            expected_type = annotations[key]
            if isinstance(expected_type, type) and issubclass(expected_type, ToolModel):
                # TODO: replace with actual foreign key
                json[key] = value["id"]                    
    return model_cls(**json)
