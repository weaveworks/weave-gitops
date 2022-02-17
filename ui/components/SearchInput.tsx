import * as React from "react";
import styled from "styled-components";
import Button from "./Button";
import Flex from "./Flex";
import Icon, { IconType } from "./Icon";
import Input from "./Input";
import Spacer from "./Spacer";

export interface Props {
  className?: string;
  setSearch: (value: string) => void;
}

const CollapsibleInput = styled(Input)`
  max-width: 0px;
  overflow: hidden;
  transition: max-width 0.5s ease;
  &.show {
    max-width: 250px;
  }
`;

function SearchInput({ className, setSearch }: Props) {
  const [show, setShow] = React.useState(false);
  return (
    <Flex className={className} align start>
      <CollapsibleInput
        className={show && "show"}
        onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
          setSearch(e.target.value)
        }
        placeholder="NAME"
      />
      <Spacer padding="xxs" />
      <Button onClick={() => setShow(!show)} variant="text" color="inherit">
        <Icon type={IconType.SearchIcon} size="medium" color="neutral30" />
      </Button>
    </Flex>
  );
}

export default styled(SearchInput).attrs({ className: SearchInput.name })``;
