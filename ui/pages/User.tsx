import * as React from "react";
import { useAuth } from "../contexts/AuthContext";

function User() {
  const { currentUser } = useAuth();

  return (
    <div>
      <div>User</div>
      <pre>{JSON.stringify(currentUser, null, 2)}</pre>
    </div>
  );
}

export default User;
