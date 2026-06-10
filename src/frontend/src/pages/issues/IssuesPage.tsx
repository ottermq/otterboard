import { useParams, useSearchParams } from "react-router-dom";
import { useIssues } from "../../hooks/useIssues";


export default function IssuesPage() {
    const { workspaceId } = useParams<{ workspaceId: string }>();
    const [searchParams, setSearchParams] = useSearchParams();
    const filters = {
        status: searchParams.get('status') ?? undefined,
        type: searchParams.get('type') ?? undefined,
        assignee: searchParams.get('assignee') ?? undefined,
        due_before: searchParams.get('due_before') ?? undefined,
        due_after: searchParams.get('due_after') ?? undefined,
        sort: searchParams.get('sort') ?? undefined,
        order: searchParams.get('order') ?? undefined,
        page: Number(searchParams.get('page') ?? 1),
        limit: 20,
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

    const { data, isLoading, error } = useIssues(workspaceId!, filters);
    if (isLoading) return <div className="p-8">Loading...</div>
    if (error) return <div className="p-8 text-red-500">Failed to load issues</div>

    return (
        <div className="p-8">
            <h1 className="text-2xl font-semibold mb-4">Issues</h1>
            <pre className="text-sm">{JSON.stringify(data, null, 2)}</pre>
        </div>
    );
}