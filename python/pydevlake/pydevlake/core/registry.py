import jsonpickle as jsonpickle
import requests as requests

from .models import PluginInfo, PluginDetails, SwaggerDoc
from .swagger import docgen


class Registry:

    def __init__(self, endpoint: str):
        self.endpoint = endpoint

    def register_plugin(self, plugin_info: PluginInfo):
        details = PluginDetails(
            plugin_info=plugin_info,
            swagger=SwaggerDoc(
                name="python",
                resource="python",
                schema=docgen.spec.to_dict()
            )
        )
        body = jsonpickle.encode(details, unpicklable=False)
        resp = requests.post(f"{self.endpoint}/plugins/register", data=body)
        if resp.status_code != 200:
            raise Exception("unexpected http status code: {}".format(resp.status_code))
