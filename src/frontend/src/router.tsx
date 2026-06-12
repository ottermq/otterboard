import { createBrowserRouter } from 'react-router-dom'
import Layout from './components/Layout'
import HomePage from './pages/Home/HomePage'
import IssuesPage from './pages/issues/IssuesPage'
import LoginPage from './pages/login/LoginPage'
import ProjectsPage from './pages/projects/ProjectsPage'

export const router = createBrowserRouter([
    { path: '/', element: <LoginPage /> },
    {
        element: <Layout />,
        children: [
            { path: '/workspaces/:workspaceId/home', element: <HomePage /> },
            { path: '/workspaces/:workspaceId/issues', element: <IssuesPage /> },
            { path: '/workspaces/:workspaceId/projects', element: <ProjectsPage /> },
            { path: '/workspaces/:workspaceId/projects/:projectId/issues', element: <IssuesPage /> },
        ]
    }
])