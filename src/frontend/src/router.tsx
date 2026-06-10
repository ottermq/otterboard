import { createBrowserRouter } from 'react-router-dom'
import IssuesPage from './pages/issues/IssuesPage'
import LoginPage from './pages/login/LoginPage'

export const router = createBrowserRouter([
    { path: '/', element: <LoginPage /> },
    { path: '/workspaces/:workspaceId/issues', element: <IssuesPage /> },
])