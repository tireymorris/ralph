import { createBrowserRouter, Navigate } from "react-router-dom";
import App from "./App";
import NewRunPage from "./pages/NewRunPage";
import RunDetail from "./pages/RunDetail";

export const router = createBrowserRouter([
  {
    path: "/",
    element: <App />,
    children: [
      { index: true, element: <NewRunPage /> },
      { path: "new", element: <Navigate to="/" replace /> },
      { path: "runs/:id", element: <RunDetail /> },
    ],
  },
]);
