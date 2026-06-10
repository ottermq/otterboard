import type { IssueDto, PaginatedResponse } from '../types'
import client from './client'

export interface IssueFilters {
    status?: string
    type?: string
    assignee?: string
    due_before?: string
    due_after?: string
    sort?: string
    order?: string
    page?: number
    limit?: number
}

export function listIssues(workspaceId: string, projectId: string, filters: IssueFilters = {}) {
    return client.get<PaginatedResponse<IssueDto>>(
        `/workspaces/${workspaceId}/projects/${projectId}/issues`,
        { params: filters },
    ).then(response => response.data)
}

export function listIssuesByWorkspace(workspaceId: string, filters: IssueFilters & { project?: string } = {}) {
    return client.get<PaginatedResponse<IssueDto>>(
        `/workspaces/${workspaceId}/issues`,
        { params: filters },
    ).then(response => response.data)
}

export function createIssue(workspaceId: string, projectId: string, input: {
    title: string
    overview?: string
    type: string
    assignee_id?: string
    due_date?: string
}) {
    return client.post<IssueDto>(`/workspaces/${workspaceId}/projects/${projectId}/issues`, input)
        .then(response => response.data)
}

export function updateIssue(workspaceId: string, projectId: string, issueId: string, input: {
    title?: string
    overview?: string
    type?: string
    status?: string
    position?: number
    assignee_id?: string
    due_date?: string
}) {
    return client.patch<IssueDto>(`/workspaces/${workspaceId}/projects/${projectId}/issues/${issueId}`, input)
        .then(response => response.data)
}

export function deleteIssue(workspaceId: string, projectId: string, issueId: string) {
    return client.delete(`/workspaces/${workspaceId}/projects/${projectId}/issues/${issueId}`)
        .then(response => response.data)
}