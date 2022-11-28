from time import sleep

import jsonpickle
import requests as requests

from circleci_models import *
from pydevlake.core.api import PluginTask
from pydevlake.core.context import Context
from pydevlake.core.ipc import plugin_method, convert, state, plugin_class

plugin_info = PluginInfo(
    name="circle_ci",
    description="circle-ci plugin",
    extension="metric",
    connection=Connection(
        table="_tool_github_connections",
    ),
    plugin_path="",
    api_endpoints=[
        ApiEndpoint(
            resource="test",
            handler="test_connection",
            method="POST"
        ),
    ],
    subtask_metas=[
        SubtaskMeta(
            name="CircleCiCollectData",
            entry_point_name="Collect_Data",
            required=True,
            enabled_by_default=True,
            description="desc",
            domain_types=[]
        ),
        SubtaskMeta(
            name="CircleCiExtractData",
            entry_point_name="Extract_Data",
            required=True,
            enabled_by_default=True,
            description="desc",
            domain_types=[]
        )
    ]
)


@plugin_class
class CircleCIApi(object):
    @plugin_method(api_doc=ApiDoc("circle_ci/test", """
        test Connection.
        ---
        post:
          responses:
            200:
              content:
                application/json:
                  schema: "{}"
    """, [Connection]))
    def test_connection(self, ctx: dict, input: ApiParamsInput) -> ApiParamsOutput:
        print("create_connection called")
        return ApiParamsOutput()


@plugin_class
class CircleCISubtasks(object):
    @plugin_method()
    def Collect_Data(self, ctx: Context):
        # import pydevlake.keon.debugger
        a: dict = ctx.settings.__dict__
        b = a['db_url']
        for i in range(1, 10):
            if state.is_cancelled:
                print("terminating")
                state.terminate()
            print('making collection progress...', flush=True)
            sleep(2)
            yield RemoteProgress(increment=10)  # the 'incremental progress'
        pass

    @plugin_method()
    def Extract_Data(self, ctx: Context):
        a: dict = ctx.settings.__dict__
        b = a['db_url']
        for i in range(1, 5):
            if state.is_cancelled:
                print("terminating")
                state.terminate()
            print('making extraction progress...', flush=True)
            sleep(2)
            yield RemoteProgress(increment=20)  # the 'incremental progress'
        pass


class CircleCiTaskOptions:
    connectionID: str


@plugin_class
class CircleCITask(PluginTask):

    @plugin_method()
    def RunMigrations(self, ctx: Context, force: bool):
        pass

    @plugin_method()
    def PrepareTaskData(self, ctx: Context, opts: CircleCiTaskOptions):
        conn_id = opts.connectionID
        body = jsonpickle.encode(ApiParamsInput(), unpicklable=False)
        resp = requests.get("http://localhost:8089/plugins/circle_ci/connections/{}".format(conn_id),
                            data=body)  # TODO endpoint must be configurable
        if resp.status_code != 200:
            raise Exception("unexpected http status code: {}".format(resp.status_code))
        conn: Connection = convert(resp.content)
        print('prepare_task_data: -> ep: {}, token: {}'.format(conn.endpoint, conn.token))
