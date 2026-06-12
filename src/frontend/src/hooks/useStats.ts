import { useQuery } from "@tanstack/react-query";
import { getWorkspaceStats } from "../api/stats";

export function useStats(workspaceId: string) {
    return useQuery({
        queryKey: ['stats', workspaceId],
        queryFn: () => getWorkspaceStats(workspaceId),
        enabled: !!workspaceId
    })
}
