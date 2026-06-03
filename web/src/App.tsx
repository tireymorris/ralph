import { NavLink, Outlet } from "react-router-dom";
import RunsList from "./components/RunsList";

export default function App() {
  return (
    <div className="app-shell">
      <aside className="app-sidebar" aria-label="Runs navigation">
        <p className="app-brand">Ralph</p>
        <nav className="app-nav" aria-label="Main">
          <NavLink to="/" end>
            Home
          </NavLink>
          <NavLink to="/new">New run</NavLink>
        </nav>
        <RunsList />
      </aside>
      <main className="app-main">
        <Outlet />
      </main>
    </div>
  );
}
