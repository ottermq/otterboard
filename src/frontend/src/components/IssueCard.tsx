import { useDraggable } from "@dnd-kit/core";
import { CSS } from "@dnd-kit/utilities";
import type { IssueDto } from "../types";

const TYPE_COLORS: Record<string, string> = {
    bug: 'text-red-600 bg-red-50',
    task: 'text-blue-600 bg-blue-50',
    story: 'text-green-600 bg-green-50',
    epic: 'text-purple-600 bg-purple-50',
}

export default function IssueCard({ issue }: { issue: IssueDto }) {
    const { attributes, listeners, setNodeRef, transform, isDragging } = useDraggable({ id: issue.id })

    return (
        <div
            ref={setNodeRef}
            style={{ transform: CSS.Translate.toString(transform), opacity: isDragging ? 0.4 : 1 }}
            {...listeners}
            {...attributes}
            className="bg-white rounded-lg p-3 shadow-sm border border-gray-200 cursor-grab active:cursor-grabbing select-none"
        >
            <p className="text-sm text-gray-900 font-medium leading-snug mb-2">{issue.title}</p>
            <div className="flex items-center gap-2">
                <span className={`text-xs px-1.5 py-0.5 rounded font-medium ${TYPE_COLORS[issue.type] ?? 'text-gray-600 bg-gray-50'}`}>
                    {issue.type}
                </span>
                {issue.due_date && (
                    <span className="text-xs text-gray-400">
                        {new Date(issue.due_date).toLocaleDateString()}
                    </span>
                )}
            </div>
        </div>
    )
}