import { useQuery } from '@tanstack/react-query';
import { listProjects } from '../api/projects';

export function useProjects(workspaceId: string) {
    return useQuery({
        queryKey: ['projects', workspaceId],
        queryFn: () => listProjects(workspaceId),
        enabled: !!workspaceId,
    });
}