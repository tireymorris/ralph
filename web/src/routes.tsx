import { createBrowserRouter } from "react-router-dom";
import App from "./App";
import NewRunPage from "./pages/NewRunPage";
import HomePage from "./pages/HomePage";
import RunDetail from "./pages/RunDetail";

export const router = createBrowserRouter([
  {
    path: "/",
    element: <App />,
    children: [
      { index: true, element: <HomePage /> },
      { path: "new", element: <NewRunPage /> },
      { path: "runs/:id", element: <RunDetail /> },
    ],
  },
]);
