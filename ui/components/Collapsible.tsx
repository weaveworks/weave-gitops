import { Collapse } from "@material-ui/core";
import React from "react";
import Icon, { IconType } from "./Icon";

const Collapsible = ({ children }) => {
  const [isOpen, setIsOpen] = React.useState(false);

  const toggle = () => setIsOpen(!isOpen);

  return (
    <div>
      <Collapse in={isOpen} collapsedSize={40}>
        {children}
      </Collapse>
      <div
        onClick={toggle}
        style={{
          display: "flex",
          justifyContent: "center",
          cursor: "pointer",
        }}
      >
        <div
          style={{
            background: "#0e0e0eb0",
            padding: "4px 8px",
            borderRadius: "8px",
            color: "white",
          }}
        >
          <Icon
            type={
              isOpen
                ? IconType.ArrowUpwardRoundedIcon
                : IconType.ArrowDownwardRoundedIcon
            }
            size="small"
            fontSize="small"
            text={isOpen ? "Collapse" : "Click to expand"}
          />
        </div>
      </div>
    </div>
  );
};

export default Collapsible;
