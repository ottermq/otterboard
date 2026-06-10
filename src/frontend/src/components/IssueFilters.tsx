const STATUS_OPTIONS = [
    { value: 'backlog', label: 'Backlog' },
    { value: 'todo', label: 'To Do' },
    { value: 'in_progress', label: 'In Progress' },
    { value: 'in_review', label: 'In Review' },
    { value: 'done', label: 'Done' },
]

const TYPE_OPTIONS = [
    { value: 'bug', label: 'Bug' },
    { value: 'task', label: 'Task' },
    { value: 'story', label: 'Story' },
    { value: 'epic', label: 'Epic' },
]

interface Props {
    status?: string
    type?: string
    dueBefore?: string
    dueAfter?: string
    onFilter: (key: string, value: string | undefined) => void
}

export default function IssueFilters({ status, type, dueBefore, dueAfter, onFilter }: Props) {
    const selectClass = "border border-gray-200 rounded px-3 py-1.5 text-sm bg-white text - gray - 700 focus: outline - none focus: ring - 2 focus: ring - blue - 500"
    const inputClass = "border border-gray-200 rounded px-3 py-1.5 text-sm bg-white text - gray - 700 focus: outline - none focus: ring - 2 focus: ring - blue - 500"

    return (
        <div className="flex flex-wrap gap-3 mb-4">
            <select
                value={status ?? ''}
                onChange={e => onFilter('status', e.target.value || undefined)}
                className={selectClass}
            >
                <option value="">All statuses</option>
                {STATUS_OPTIONS.map(o => (
                    <option key={o.value} value={o.value}>{o.label}</option>
                ))}
            </select>

            <select
                value={type ?? ''}
                onChange={e => onFilter('type', e.target.value || undefined)}
                className={selectClass}
            >
                <option value="">All types</option>
                {TYPE_OPTIONS.map(o => (
                    <option key={o.value} value={o.value}>{o.label}</option>
                ))}
            </select>

            <input
                type="date"
                value={dueAfter ?? ''}
                onChange={e => onFilter('due_after', e.target.value || undefined)}
                className={inputClass}
                title="Due after"
            />

            <input
                type="date"
                value={dueBefore ?? ''}
                onChange={e => onFilter('due_before', e.target.value || undefined)}
                className={inputClass}
                title="Due before"
            />

            {(status || type || dueBefore || dueAfter) && (
                <button
                    onClick={() => {
                        onFilter('status', undefined)
                        onFilter('type', undefined)
                        onFilter('due_before', undefined)
                        onFilter('due_after', undefined)
                    }}
                    className="text-sm text-gray-500 hover:text-gray-800 underline"
                >
                    Clear filters
                </button>
            )}
        </div>
    )
}