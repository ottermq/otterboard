import { useQuery } from '@tanstack/react-query';
import { listIssues, listIssuesByWorkspace, type IssueFilters } from '../api/issues';

export function useIssues(
    workspaceId: string,
    filters: IssueFilters & { project?: string } = {},
    projectId?: string
) {
    return useQuery({
        queryKey: ['issues', workspaceId, projectId, filters],
        queryFn: () => projectId
            ? listIssues(workspaceId, projectId, filters)
            : listIssuesByWorkspace(workspaceId, filters),
        enabled: !!workspaceId,
    });
}

