import { useQuery } from '@tanstack/react-query';
import { listIssuesByWorkspace, type IssueFilters } from '../api/issues';

export function useIssues(workspaceId: string, filters: IssueFilters & { project?: string } = {}) {
    return useQuery({
        queryKey: ['issues', workspaceId, filters],
        queryFn: () => listIssuesByWorkspace(workspaceId, filters),
        enabled: !!workspaceId,
    });
}

