import { Link, useParams } from 'react-router-dom';
import { useProjects } from '../../hooks/useProjecs';

export default function ProjectsPage() {
    const { workspaceId } = useParams<{ workspaceId: string }>();
    const { data: projects, isLoading, isError } = useProjects(workspaceId!);

    if (isLoading) return <div className="p-8 text-gray-500">Loading...</div>
    if (isError) return <div className="p-8 text-red-500">Failed to load projects.</div>

    return (
        <div className="p-8 max-w-4xl mx-auto">
            <h1 className="text-2xl font-semibold mb-6">Projects</h1>
            {projects?.length === 0 ? (
                <p className="text-gray-400">No projects yet.</p>
            ) : (
                <div className="grid gap-3">
                    {projects?.map(project => (
                        <Link
                            key={project.id}
                            to={`/workspaces/${workspaceId}/projects/${project.id}/issues`}
                            className="block p-4 border border-gray-200 rounded-lg hover:bg-gray-50"
                        >
                            <span className="font-medium text-gray-900">{project.name}</span>
                        </Link>
                    ))}
                </div>
            )}
        </div>
    )
}
