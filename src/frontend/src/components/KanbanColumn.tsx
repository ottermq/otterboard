import { useDroppable } from "@dnd-kit/core";
import type { IssueDto } from "../types";
import IssueCard from "./IssueCard";

interface Props {
    status: string
    label: string
    issues: IssueDto[]
}

export default function KanbanColumn({ status, label, issues }: Props) {
    const { setNodeRef, isOver } = useDroppable({ id: status })

    return (
        <div
            ref={setNodeRef}
            className={`flex-1 min-w-4 rounded-lg p-3 flex flex-col gap-2 min-h-32 transition-colors
                ${isOver
                    ? 'bg-blue-50 border-2 border-blue-200'
                    : 'bg-gray-100'
                }`}
        >
            <div className="flex items-center justify-between px-1 mb-1">
                <span className="text-sm font-semibold text-gray-700">{label}</span>
                <span className="text-xs text-gray-400 bg-gray-200 rounded-full px-2">{issues.length}</span>
            </div>
            {issues.map(issue => (
                <IssueCard key={issue.id} issue={issue} />
            ))}
        </div>
    )
}