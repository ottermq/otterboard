import { useState } from "react";
import { useIssues } from "../hooks/useIssues";

interface Props {
    workspaceId: string
    projectId?: string
}

const WEEKDAYS = ['Sun', 'Mon', 'Tue', 'Web', 'Thu', 'Fri', 'Sat']

function pad(n: number) {
    return String(n).padStart(2, '0')
}

export default function CalendarView({ workspaceId, projectId }: Props) {
    const [current, setCurrent] = useState(() => {
        const d = new Date()
        return new Date(d.getFullYear(), d.getMonth(), 1)
    })

    const year = current.getFullYear()
    const month = current.getMonth()

    const due_after = `${year}-${pad(month + 1)}-01`
    const due_before = `${year}-${pad(month + 1)}-${pad(new Date(year, month + 1, 0).getDate())}`

    const { data } = useIssues(workspaceId, { due_after, due_before, limit: 100 }, projectId)
    const issues = data?.data ?? []

    const firstWeekday = new Date(year, month, 1).getDay()
    const daysInMonth = new Date(year, month + 1, 0).getDate()

    const cells = Array.from({ length: 42 }, (_, i) => {
        const day = i - firstWeekday + 1
        return day >= 1 && day <= daysInMonth ? day : null
    })

    const today = new Date()
    const isToday = (day: number) =>
        today.getFullYear() === year &&
        today.getMonth() == month &&
        today.getDate() == day

    function issuesForDay(day: number) {
        const dateStr = `${year}-${pad(month + 1)}-${pad(day)}`
        return issues.filter(i => i.due_date?.startsWith(dateStr))
    }

    function prev() { setCurrent(d => new Date(d.getFullYear(), d.getMonth() - 1, 1)) }
    function next() { setCurrent(d => new Date(d.getFullYear(), d.getMonth() + 1, 1)) }

    const monthLabel = current.toLocaleString('default', { month: 'long', year: 'numeric' })

    return (
        <div>
            {/*  Month navigation */}
            <div className="flex items-center justify-center mb-4">
                <button
                    onClick={prev}
                    className="px-3 py-1 text-sm rounded hover:bg-gray-100 text-gray-600"
                >
                    ←
                </button>
                <span className="font-semibold text-gray-800">{monthLabel}</span>
                <button
                    onClick={next}
                    className="px-3 py-1 text-sm rounded hover:bg-gray-100 text-gray-600"
                >
                    →
                </button>
            </div>

            {/* Weekday headers */}
            <div className="grid grid-cols-7 mb-px">
                {WEEKDAYS.map(d => (
                    <div key={d} className="text-xs text-center text-gray-400 py-2 font-medium uppercase tracking-wide">
                        {d}
                    </div>
                ))}
            </div>

            {/*  Day grid */}
            <div className="grid grid-cols-7 gap-px bg-gray-200 border border-gray-200 rounded-lg overflow-hidden">
                {cells.map((day, i) => (
                    <div
                        key={i}
                        className={`min-h-24 p-2 ${day === null ? 'bg-gray-50' : 'bg-white'}`}
                    >
                        {day !== null && (
                            <>
                                <span className={`text-xl font-medium inline-flex items-center justify-center w-6 h-6 rounded-full
                                    ${isToday(day)
                                        ? 'bg-blue-600 text-white'
                                        : 'text-gray-600'
                                    }`}>
                                    {day}
                                </span>
                                <div className="mt-1 flex flex-col gap-0.5">
                                    {issuesForDay(day).map(issue => (
                                        <div
                                            key={issue.id}
                                            className="text-xs bg-blue-100 text-blue-800 rounded px-1.5 py-0.5 truncate"
                                            title={issue.title}
                                        >
                                            {issue.title}
                                        </div>
                                    ))}
                                </div>
                            </>
                        )}
                    </div>
                ))}
            </div>
        </div>
    )
}