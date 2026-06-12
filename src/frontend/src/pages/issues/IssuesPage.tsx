import { useState } from 'react';
import { useParams, useSearchParams } from "react-router-dom";
import CalendarView from '../../components/CalendarView';
import IssueFilters from "../../components/IssueFilters";
import IssueForm from '../../components/IssueForm';
import IssueTable from "../../components/IssueTable";
import KanbanBoard from '../../components/KanbanBoard';
import Modal from '../../components/Modal';
import Pagination from "../../components/Pagination";
import { useIssues } from "../../hooks/useIssues";
import type { IssueDto } from '../../types';

export default function IssuesPage() {
    const [view, setView] = useState<'table' | 'kanban' | 'calendar'>('table')
    const LIMIT = view === 'kanban' ? 100 : 20;

    const { workspaceId, projectId } = useParams<{ workspaceId: string; projectId?: string }>();
    const [searchParams, setSearchParams] = useSearchParams();
    const [showCreate, setShowCreate] = useState(false)
    const [editIssue, setEditIssue] = useState<IssueDto | null>(null)

    const filters = {
        status: searchParams.get('status') ?? undefined,
        type: searchParams.get('type') ?? undefined,
        assignee: searchParams.get('assignee') ?? undefined,
        due_before: searchParams.get('due_before') ?? undefined,
        due_after: searchParams.get('due_after') ?? undefined,
        sort: searchParams.get('sort') ?? undefined,
        order: searchParams.get('order') ?? undefined,
        page: Number(searchParams.get('page') ?? 1),
        limit: LIMIT,
    }

    function setFilter(key: string, value: string | undefined) {
        setSearchParams(prev => {
            const next = new URLSearchParams(prev);
            if (value) next.set(key, value);
            else next.delete(key);
            if (key !== 'page') next.delete('page');
            return next;
        });
    }

    function handleSort(key: string) {
        const currentOrder = filters.sort === key && filters.order === 'asc' ? 'desc' : 'asc';
        setSearchParams(prev => {
            const next = new URLSearchParams(prev);
            next.set('sort', key);
            next.set('order', currentOrder);
            next.delete('page');
            return next;
        });

    }

    const { data, isLoading, error } = useIssues(workspaceId!, filters, projectId);

    if (isLoading) return <div className="p-8 text-gray-500">Loading...</div>
    if (error) return <div className="p-8 text-red-500">Failed to load issues</div>

    return (
        <div className="p-8 max-w-6xl mx-auto">
            <div className="flex items-center justify-between mb-6">
                <h1 className="text-2xl font-semibold">{projectId ? 'Project Issues' : 'All Issues'}</h1>
                <div className="flex items-center gap-2">
                    <button
                        onClick={() => setView('table')}
                        className={`px-3 py-1.5 text-sm rounded 
                            ${view === 'table'
                                ? 'bg-gray-200 font-medium'
                                : 'text-gray-500 hover:bg-gray-100'
                            }`}
                    >
                        Table
                    </button>
                    <button
                        onClick={() => setView('kanban')}
                        className={`px-3 py-1.5 text-sm rounded 
                            ${view === 'kanban'
                                ? 'bg-gray-200 font-medium'
                                : 'text-gray-500 hover:bg-gray-100'
                            }`}
                    >
                        Kanban
                    </button>
                    <button
                        onClick={() => setView('calendar')}
                        className={`px-3 py-1.5 text-sm rounded 
                            ${view === 'calendar'
                                ? 'bg-gray-200 font-medium'
                                : 'text-gray-500 hover:bg-gray-100'
                            }`}
                    >
                        Calendar
                    </button>
                    {projectId && (
                        <button
                            onClick={() => setShowCreate(true)}
                            className="px-4 py-2 text-sm bg-blue-600 text-white rounded hover:bg-blue-700"
                        >
                            New Issue
                        </button>
                    )}
                </div>
            </div>
            {view === 'table' ? (
                <>
                    <IssueFilters
                        status={filters.status}
                        type={filters.type}
                        dueBefore={filters.due_before}
                        dueAfter={filters.due_after}
                        onFilter={setFilter}
                    />
                    <IssueTable
                        issues={data?.data ?? []}
                        sortBy={filters.sort}
                        sortOrder={filters.order}
                        onSort={handleSort}
                        onRowClick={projectId ? setEditIssue : undefined}
                    />
                    <Pagination
                        page={filters.page}
                        limit={LIMIT}
                        total={data?.total ?? 0}
                        onPage={page => setFilter('page', String(page))}
                    />
                </>
            ) : view === 'kanban' ? (
                <KanbanBoard
                    issues={data?.data ?? []}
                    workspaceId={workspaceId!} />
            ) : (
                <CalendarView workspaceId={workspaceId!} projectId={projectId} />
            )}
            {showCreate && projectId && (
                <Modal title="New Issue" onClose={() => setShowCreate(false)}>
                    <IssueForm
                        workspaceId={workspaceId!}
                        projectId={projectId}
                        onClose={() => setShowCreate(false)}
                    />
                </Modal>
            )}

            {editIssue && projectId && (
                <Modal title="Edit Issue" onClose={() => setEditIssue(null)}>
                    <IssueForm
                        workspaceId={workspaceId!}
                        projectId={projectId}
                        issue={editIssue}
                        onClose={() => setEditIssue(null)}
                    />
                </Modal>
            )}
        </div>
    )
}