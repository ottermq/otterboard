import { useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import { createIssue, updateIssue } from '../api/issues';
import type { IssueDto } from '../types';

const TYPE_OPTIONS = [
    { value: 'bug', label: 'Bug' },
    { value: 'task', label: 'Task' },
    { value: 'story', label: 'Story' },
    { value: 'epic', label: 'Epic' },
]

const STATUS_OPTIONS = [
    { value: 'backlog', label: 'Backlog' },
    { value: 'todo', label: 'To Do' },
    { value: 'in_progress', label: 'In Progress' },
    { value: 'in_review', label: 'In Review' },
    { value: 'done', label: 'Done' },
]

const inputClass = "w-full border border-gray-200 rounded px-3 py-2 text-sm  focus:outline-none focus:ring-2 focus:ring-blue-500"
const selectClass = "w-full border border-gray-200 rounded px-3 py-2 text-sm bg-white focus:outline-none focus:ring-2 focus:ring-blue-500"
const labelClass = "block text-sm font-medium text-gray-700 mb-1"

interface Props {
    workspaceId: string
    projectId: string
    issue?: IssueDto
    onClose: () => void
}

export default function IssueForm({ workspaceId, projectId, issue, onClose }: Props) {
    const queryClient = useQueryClient()
    const isEdit = !!issue

    const [title, setTitle] = useState(issue?.title ?? '')
    const [overview, setOverview] = useState(issue?.overview ?? '')
    const [type, setType] = useState(issue?.type ?? 'task')
    const [status, setStatus] = useState(issue?.status ?? 'backlog')
    const [assigneeId, setAssigneeId] = useState(issue?.assignee_id ?? '')
    const [dueDate, setDueDate] = useState(issue?.due_date ?? '')
    const [error, setError] = useState('')
    const [loading, setLoading] = useState(false)

    async function handleSubmit(e: React.FormEvent) {
        e.preventDefault()
        setError('')
        setLoading(true)
        try {
            if (isEdit) {
                await updateIssue(workspaceId, projectId, issue.id, {
                    title,
                    overview: overview || undefined,
                    type,
                    status,
                    assignee_id: assigneeId || undefined,
                    due_date: dueDate || null,
                    position: issue.position,
                })
            } else {
                await createIssue(workspaceId, projectId, {
                    title,
                    overview: overview || undefined,
                    type,
                    assignee_id: assigneeId || undefined,
                    due_date: dueDate || undefined,
                })
            }
            await queryClient.invalidateQueries({ queryKey: ['issues', workspaceId] })
            onClose()
        } catch {
            setError('Something went wrong. Please try again.')
        } finally {
            setLoading(false)
        }
    }
    return (
        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
            {error && <p className="text-red-500 text-sm">{error}</p>}

            <div>
                <label className={labelClass}>Title <span className="text-red-500">*</span></label>
                <input
                    type="text"
                    value={title}
                    onChange={e => setTitle(e.target.value)}
                    className={inputClass}
                    required
                />
            </div>

            <div>
                <label className={labelClass}>Overview</label>
                <textarea
                    value={overview}
                    onChange={e => setOverview(e.target.value)}
                    className={inputClass}
                    rows={3}
                />
            </div>

            <div className="grid grid-cols-2 gap-4">
                <div>
                    <label className={labelClass}>Type <span className="text-red-500">*</span></label>
                    <select value={type} onChange={e => setType(e.target.value)}
                        className={selectClass}>
                        {TYPE_OPTIONS.map(o => <option key={o.value} value={o.value}>{o.label}</option>)}
                    </select>
                </div>
                {isEdit && (
                    <div>
                        <label className={labelClass}>Status</label>
                        <select value={status} onChange={e => setStatus(e.target.value)}
                            className={selectClass}>
                            {STATUS_OPTIONS.map(o => <option key={o.value}
                                value={o.value}>{o.label}</option>)}
                        </select>
                    </div>
                )}
            </div>

            <div className="grid grid-cols-2 gap-4">
                <div>
                    <label className={labelClass}>Assignee ID</label>
                    <input
                        type="text"
                        value={assigneeId}
                        onChange={e => setAssigneeId(e.target.value)}
                        className={inputClass}
                        placeholder="UUID"
                    />
                </div>
                <div>
                    <label className={labelClass}>Due Date</label>
                    <input
                        type="date"
                        value={dueDate}
                        onChange={e => setDueDate(e.target.value)}
                        className={inputClass}
                    />
                </div>
            </div>

            <div className="flex justify-end gap-3 pt-2">
                <button type="button" onClick={onClose} className="px-4 py-2 text-sm text-gray-600
  hover:text-gray-800">
                    Cancel
                </button>
                <button
                    type="submit"
                    disabled={loading}
                    className="px-4 py-2 text-sm bg-blue-600 text-white rounded hover:bg-blue-700
  disabled:opacity-50"
                >
                    {loading ? 'Saving...' : isEdit ? 'Save changes' : 'Create issue'}
                </button>
            </div>
        </form>
    )
}