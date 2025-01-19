import * as React from "react";
import styled from "styled-components";
import { IconButton } from "./Button";
import Flex from "./Flex";
import Icon, { IconType } from "./Icon";
import Input from "./Input";

type Props = {
  className?: string;
  onSubmit: (val: string) => void;
};

const Expander = styled(
  ({
    expanded,
    className,
    children,
  }: {
    expanded: boolean;
    className?: string;
    children?: any;
  }) => (
    <div className={`${className} ${expanded ? "expanded" : ""}`}>
      {children}
    </div>
  ),
)`
  width: 0px;
  transition-property: opacity, width;
  transition-duration: 0.5s;
  transition-timing-function: cubic-bezier(0.46, 0.03, 0.52, 0.96);
  margin-left: 4px;
  opacity: 0;

  &.expanded {
    width: 200px;
    opacity: 1;
  }

  input {
    padding: 8px 10px;
    border-bottom: 1px solid ${(props) => props.theme.colors.neutral40};
  }
`;

function SearchField({ className, onSubmit }: Props) {
  const inputRef = React.createRef<HTMLInputElement>();
  const [expanded, setExpanded] = React.useState(false);
  const [value, setValue] = React.useState("");

  const handleExpand = (ev) => {
    ev.preventDefault();

    if (!expanded) {
      inputRef.current.focus();
    } else {
      inputRef.current.blur();
    }
    setValue("");
    setExpanded(!expanded);
  };

  const handleSubmit = (e) => {
    e.preventDefault();
    setValue("");
    onSubmit(value);
  };

  return (
    <Flex align className={className}>
      <IconButton
        onClick={handleExpand}
        className={className}
        variant="text"
        color="inherit"
        size="large"
      >
        <Icon
          type={IconType.ExploreIcon}
          size="medium"
          color={expanded ? "primary" : "neutral30"}
        />
      </IconButton>
      <Expander expanded={expanded}>
        <form onSubmit={handleSubmit}>
          <Input
            id="table-search"
            placeholder="Search"
            inputRef={inputRef}
            value={value}
            onChange={(ev) => setValue(ev.target.value)}
          />
        </form>
      </Expander>
    </Flex>
  );
}

export default styled(SearchField).attrs({ className: SearchField.name })``;
