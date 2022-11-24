from typing import Iterable, Tuple, Optional

from pydevlake import Stream
from pydevlake.domain_layer.devops import CICDPipeline

from models import Pipeline


class Pipelines(Stream):
    tool_model = Pipeline
    domain_model = CICDPipeline

    def collect(self, state) -> Iterable[Tuple[object, dict]]:
        for json in self.plugin.api.pipelines(self.conf['project_slug']):
            yield json, state
    
    def convert(self, pipeline: Pipeline) -> Iterable[CICDPipeline]:
        yield CICDPipeline(
            status=self.convert_status(pipeline),
            created_date=pipeline.created_at,
            finished_date=pipeline.updated_at,
            name=f'{pipeline.project_slug}:{pipeline.number}',
            result=self.convert_result(pipeline),
            duration_sec=(pipeline.updated_at - pipeline.created_at).seconds if pipeline.created_at else None
        )

    def convert_status(self, pipeline: Pipeline) -> CICDPipeline.Status:
        match pipeline.state:
            case Pipeline.State.CREATED | Pipeline.State.ERRORED:
                return CICDPipeline.Status.DONE
            case _:
                return CICDPipeline.Status.IN_PROGRESS

    def convert_result(self, pipeline: Pipeline) -> Optional[CICDPipeline.Result]:
        match pipeline.state:
            case Pipeline.State.CREATED:
                return CICDPipeline.Result.SUCCESS
            case Pipeline.State.ERRORED:
                return CICDPipeline.Result.FAILURE
            case _:
                return None