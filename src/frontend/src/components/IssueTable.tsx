import type { IssueDto } from "../types"

const STATUS_LABELS: Record<string, string> = {
    backlog: 'Backlog',
    todo: 'To Do',
    in_progress: 'In Progress',
    in_review: 'In Review',
    done: 'Done',
}

const TYPE_LABELS: Record<string, string> = {
    bug: 'Bug',
    task: 'Task',
    story: 'Story',
    epic: 'Epic',
}

interface Column {
    key: string
    label: string
    sortable: boolean
}

const COLUMNS: Column[] = [
    { key: 'title', label: 'Title', sortable: true },
    { key: 'type', label: 'Type', sortable: true },
    { key: 'status', label: 'Status', sortable: true },
    { key: 'due_date', label: 'Due Date', sortable: true },
    { key: 'created_at', label: 'Created', sortable: true },
]

interface Props {
    issues: IssueDto[]
    sortBy?: string
    sortOrder?: string
    onSort: (key: string) => void
    onRowClick?: (issue: IssueDto) => void
}

export default function IssueTable({ issues, sortBy, sortOrder, onSort, onRowClick }: Props) {
    function SortIndicator({ column }: { column: string }) {
        if (sortBy !== column) return <span className="ml-1 text-gray-300">⇅</span>
        return <span className="ml-1">{sortOrder === 'desc' ? '↓' : '↑'}</span>
    }

    return (
        <div className="overflow-x-auto rounded-lg border border-gray-200">
            <table className="w-full text-sm text-left">
                <thead className="bg-gray-50 text-gray-600 uppercase text-xs">
                    <tr>
                        {COLUMNS.map(col => (
                            <th key={col.key} className="px-4 py-3 font-medium">
                                {col.sortable ? (
                                    <button
                                        onClick={() => onSort(col.key)}
                                        className="flex items-center hover:text-gray-900"
                                    >
                                        {col.label}
                                        <SortIndicator column={col.key} />
                                    </button>
                                ) : col.label}
                            </th>
                        ))}
                    </tr>
                </thead>
                <tbody className="divide-y divide-gray-100">
                    {issues.length === 0 ? (
                        <tr>
                            <td colSpan={COLUMNS.length} className="px-4 py-8 text-center text-gray-400">
                                No issues found.
                            </td>
                        </tr>
                    ) : issues.map(issue => (
                        <tr
                            key={issue.id}
                            onClick={() => onRowClick?.(issue)}
                            className={`hover:bg-gray-50 ${onRowClick ? 'cursor-pointer' : ''}`}>
                            <td className="px-4 py-3 font-medium text-gray-900">{issue.title}</td>
                            <td className="px-4 py-3 text-gray-600">{TYPE_LABELS[issue.type] ??
                                issue.type}</td>
                            <td className="px-4 py-3 text-gray-600">{STATUS_LABELS[issue.status] ??
                                issue.status}</td>
                            <td className="px-4 py-3 text-gray-600">
                                {issue.due_date ? new Date(issue.due_date).toLocaleDateString() : '—'}
                            </td>
                            <td className="px-4 py-3 text-gray-600">
                                {new Date(issue.created_at).toLocaleDateString()}
                            </td>
                        </tr>
                    ))}
                </tbody>
            </table>
        </div>
    )
}