import client from "./client";


export interface ProjectDto {
    id: string
    workspace_id: string
    name: string
    image_url: string
    created_at: string
    updated_at: string
}

export interface CreateProjectInput {
    name: string
    image_url?: string
}

export interface UpdateProjectInput {
    name?: string
    image_url?: string
}

export function listProjects(workspaceId: string) {
    return client.get<ProjectDto[]>(`/workspaces/${workspaceId}/projects`)
        .then(response => response.data)
}

export function createProject(workspaceId: string, input: CreateProjectInput) {
    return client.post<ProjectDto>(`/workspaces/${workspaceId}/projects`, input)
        .then(response => response.data)
}

export function updateProject(workspaceId: string, projectId: string, input: UpdateProjectInput) {
    return client.patch<ProjectDto>(`/workspaces/${workspaceId}/projects/${projectId}`, input)
        .then(response => response.data)
}

export function deleteProject(workspaceId: string, projectId: string) {
    return client.delete(`/workspaces/${workspaceId}/projects/${projectId}`)
        .then(response => response.data)
}
