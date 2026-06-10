import client from "./client";

export interface WorkspaceDto {
    id: string;
    name: string;
    owner_id: string;
    created_at: string;
    updated_at: string;
}

export function listWorkspaces() {
    return client.get<WorkspaceDto[]>('/workspaces').then(response => response.data);
}
