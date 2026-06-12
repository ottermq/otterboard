import { useParams } from "react-router-dom"
import IssueTable from "../../components/IssueTable"
import { useIssues } from "../../hooks/useIssues"
import { useStats } from "../../hooks/useStats"

const STAT_CARDS = [
    { label: 'Total Projects', key: 'total_projects' },
    { label: 'Total Issues', key: 'total_issues' },
    { label: 'Assigned to Me', key: 'assigned_issues' },
    { label: 'Completed', key: 'completed_issues' },
    { label: 'Overdue', key: 'overdue_issues' },
] as const

export default function HomePage() {
    const { workspaceId } = useParams<{ workspaceId: string }>()

    const { data: stats, isLoading: statsLoading } = useStats(workspaceId!)
    const { data: issues, isLoading: issuesLoading } = useIssues(workspaceId!, { assignee: 'me', limit: 20 })

    return (
        <div className="p-8 max-w-6xl mx-auto">
            <h1 className="text-2xl font-semibold mb-6">Home</h1>

            {/* Stats row */}
            <div className="grid grid-cols-5 gap-4 mb-8">
                {STAT_CARDS.map(({ label, key }) => (
                    <div key={key} className="bg-white border border-gray-200 rounded-lg p-4">
                        <p className="text-xs text-gray-500 mb-1">{label}</p>
                        <p className="text-2xl font-bold text-gray-900">
                            {statsLoading ? '-' : (stats?.[key] ?? 0)}
                        </p>
                    </div>
                ))}
            </div>

            {/* Assigned issues */}
            <h2 className="text-lg font-semibold mb-3">My Issues</h2>
            {issuesLoading
                ? <p className="text-gray-500 text-sm">Loading...</p>
                : <IssueTable issues={issues?.data ?? []} />
            }
        </div>
    )
}