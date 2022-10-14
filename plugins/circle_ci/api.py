from pydevlake.api import API, TokenPaginator, request_hook


class CircleCIAPI(API):
    base_url = 'https://circleci.com/api/v2/'

    paginator = TokenPaginator(
        items_attr='items',
        next_page_token_attr='next_page_token',
        next_page_token_param='page-token'
    )

    @request_hook
    def kebab_case_query_params(self, request):
        """
        CircleCI API uses kebab-cased query parameters
        that are not valid python identifier.
        So we need to convert from snake_cased to kebab-cased.  
        """
        request.query_args = {
            key.replace('_', '-'): value
            for key, value 
            in request.query_args.items()
        }

    def pipelines(self, project_slug: str):
        return self.get(f'project/{project_slug}/pipeline')

    def pipeline_workflows(self, pipeline_id: str):
        return self.get(f'pipeline/{pipeline_id}/workflow')

    def workflow_jobs(self, workflow_id: str):
        return self.get(f'workflow/{workflow_id}/job')
