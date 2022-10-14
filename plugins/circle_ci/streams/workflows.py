from typing import Iterable, Tuple, Optional

from pydevlake import Substream
from pydevlake.domain_layer.devops import CICDPipeline

from models import Workflow, Pipeline
from streams.pipelines import Pipelines


class Workflows(Substream):
    parent_stream = Pipelines
    tool_model =  Workflow
    domain_model = CICDPipeline

    def collect(self, state: dict, pipeline: Pipeline) -> Iterable[Tuple[object, dict]]:
        for json in self.plugin.api.pipeline_workflows(pipeline.id):
            yield json, state
    
    def convert(self, workflow: Workflow) -> Iterable[CICDPipeline]:
        yield CICDPipeline(
            status=self.convert_status(workflow),
            created_date=workflow.created_at,
            finished_date=workflow.stopped_at,
            name = workflow.name,
            result = self.convert_result(workflow),
            duration_sec = (workflow.stopped_at - workflow.stopped_at).seconds if workflow.stopped_at else None
        )

    def convert_status(self, workflow: Workflow) -> CICDPipeline.Status:
        s = Workflow.Status
        match workflow.status:
            case s.RUNNING | s.FAILING | s.ON_HOLD:
                return CICDPipeline.Status.IN_PROGRESS
            case _:
                return CICDPipeline.Status.DONE

    def convert_result(self, workflow: Workflow) -> Optional[CICDPipeline.Result]:
        s = Workflow.Status
        match workflow.status:
            case s.SUCCESS:
                return CICDPipeline.Result.SUCCESS
            case s.FAILED | s.ERROR | s.UNAUTHORIZED:
                return CICDPipeline.Result.FAILURE
            case s.CANCELED | s.NOT_RUN:
                return CICDPipeline.Result.ABORT
            case _:
                return None
