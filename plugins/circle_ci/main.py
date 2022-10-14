from pydevlake import Plugin

from streams.pipelines import Pipelines
from streams.workflows import Workflows
from api import CircleCIAPI


class CircleCIPlugin(Plugin):
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
    CircleCIPlugin.cli()
