import { createRootRoute, Link, Outlet } from "@tanstack/react-router";
import "./navbar.css";

export const Route = createRootRoute({
  component: () => (
    <>
      <div className="flex gap-2 p-4 text-white border-b navbar-bg">
        <Link to="/" className="[&.active]:font-bold active-item">
          Home
        </Link>
      </div>
      <Outlet />
    </>
  ),
});
