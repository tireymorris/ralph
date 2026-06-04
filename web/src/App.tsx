import { Outlet } from "react-router-dom";
import UpdateButton from "./components/UpdateButton";

export default function App() {
  return (
    <main className="app-main">
      <header className="app-topbar">
        <UpdateButton />
      </header>
      <div className="app-content">
        <Outlet />
      </div>
    </main>
  );
}
