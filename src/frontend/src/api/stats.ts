import client from "./client";

export interface StatsDto {
    total_projects: number
    total_issues: number
    assigned_issues: number
    completed_issues: number
    overdue_issues: number
}

export function getWorkspaceStats(workspaceId: string): Promise<StatsDto> {
    return client.get<StatsDto>(`/workspaces/${workspaceId}/stats`)
        .then(response => response.data)
}