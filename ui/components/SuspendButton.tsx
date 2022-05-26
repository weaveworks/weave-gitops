import * as React from "react";
import styled from "styled-components";
import Button from "./Button";

type Props = {
  className?: string;
  loading?: boolean;
  toggleSuspend: (req) => void;
  suspend: boolean;
};

function SuspendButton({ className, loading, toggleSuspend, suspend }: Props) {
  return (
    <Button
      className={className}
      variant="outlined"
      onClick={toggleSuspend}
      loading={loading}
    >
      {suspend ? "Resume" : "Suspend"}
    </Button>
  );
}

export default styled(SuspendButton).attrs({ className: SuspendButton.name })``;
