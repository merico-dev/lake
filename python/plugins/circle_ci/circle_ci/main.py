from api import CircleCIAPI
from pydevlake.plugin import Plugin
from pydevlake import startup
from streams.pipelines import Pipelines
from streams.workflows import Workflows

from circleci_plugin import CircleCIPlugin


class CircleCIPlugin2(Plugin):
    def __init__(self, **conf):
        super().__init__(**conf)
        self.api = CircleCIAPI()

    @property
    def streams(self):
        return [
            Pipelines(self),
            Workflows(self)
        ]


if __name__ == '__main__':
    # CircleCIPlugin2.cli() #TODO adapt this to the AbstractPlugin base class
    print("starting circle_ci")
    startup.init_cmd(__file__, CircleCIPlugin())
    # pass
