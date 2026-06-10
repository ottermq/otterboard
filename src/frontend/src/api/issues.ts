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