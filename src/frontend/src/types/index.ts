export interface IssueDto {
    id: string;
    project_id: string;
    title: string;
    overview: string;
    type: string;
    status: string;
    position: number;
    assignee_id: string;
    created_by: string;
    due_date: string | null;
    created_at: string;
    updated_at: string;
}

export interface PaginatedResponse<T> {
    data: T[];
    total: number;
    page: number;
    limit: number;
}
