import { DndContext, type DragEndEvent } from "@dnd-kit/core";
import { useQueryClient } from "@tanstack/react-query";
import { useEffect, useState } from "react";
import { updateIssue } from "../api/issues";
import type { IssueDto } from "../types";
import KanbanColumn from "./KanbanColumn";

const COLUMNS = [
    { key: 'backlog', label: 'Backlog' },
    { key: 'todo', label: 'To Do' },
    { key: 'in_progress', label: 'In Progress' },
    { key: 'in_review', label: 'In Review' },
    { key: 'done', label: 'Done' },
]

interface Props {
    issues: IssueDto[]
    workspaceId: string
}

export default function KanbanBoard({ issues, workspaceId }: Props) {
    const queryClient = useQueryClient()
    const [local, setLocal] = useState(issues)

    useEffect(() => { setLocal(issues) }, [issues])

    async function handleDragEnd(event: DragEndEvent) {
        const { active, over } = event
        if (!over) return

        const issueId = active.id as string
        const newStatus = over.id as string
        const issue = local.find(i => i.id === issueId)
        if (!issue || issue.status === newStatus) return

        // optimistic update
        setLocal(prev => prev.map(i => i.id === issueId ? { ...i, status: newStatus } : i))

        try {
            await updateIssue(workspaceId, issue.project_id, issueId, {
                title: issue.title,
                overview: issue.overview || undefined,
                type: issue.type,
                status: newStatus,
                position: issue.position,
                assignee_id: issue.assignee_id || undefined,
                due_date: issue.due_date || null,
            })
            queryClient.invalidateQueries({ queryKey: ['issues', workspaceId] })
        } catch {
            setLocal(prev => prev.map(i => i.id === issueId ? { ...i, status: issue.status } : i))
        }
    }

    return (
        <DndContext onDragEnd={handleDragEnd}>
            <div className="flex gap-4 overflow-x-auto pb4">
                {COLUMNS.map(col => (
                    <KanbanColumn
                        key={col.key}
                        status={col.key}
                        label={col.label}
                        issues={local.filter(i => i.status == col.key)}
                    />
                ))}
            </div>
        </DndContext>
    )

}