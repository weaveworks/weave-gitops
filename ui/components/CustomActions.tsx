import * as React from "react";
import Spacer from "./Spacer";

function CustomActions({ actions }: { actions: JSX.Element[] }) {
  return actions?.length > 0 ? (
    <>
      {actions?.map((action, index) => (
        <React.Fragment key={index}>
          <Spacer padding="xs" />
          {action}
        </React.Fragment>
      ))}
    </>
  ) : null;
}

export default CustomActions;
