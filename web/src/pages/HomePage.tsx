import { Link } from "react-router-dom";

export default function HomePage() {
  return (
    <section className="home-hero">
      <h1>Ralph</h1>
      <p>
        Turn a goal into a working implementation. Ralph writes a PRD, walks
        you through review, and implements each story with an AI coding
        backend.
      </p>
      <Link to="/new" className="btn btn--primary home-hero-cta">
        Start a new run
      </Link>
    </section>
  );
}
