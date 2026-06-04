import { Outlet } from "react-router-dom";
import CleanButton from "./components/CleanButton";
import UpdateButton from "./components/UpdateButton";

export default function App() {
  return (
    <main className="app-main">
      <header className="app-topbar">
        <CleanButton />
        <UpdateButton />
      </header>
      <div className="app-content">
        <Outlet />
      </div>
    </main>
  );
}
