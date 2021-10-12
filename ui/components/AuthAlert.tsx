import * as React from "react";
import styled from "styled-components";
import { RequestError } from "../lib/types";
import Alert from "./Alert";
import Button from "./Button";
import Flex from "./Flex";

type Props = {
  className?: string;
  error?: RequestError;
  onClick: () => void;
  title: string;
};

const Message = ({ onClick }) => (
  <Flex align wide between>
    Could not authenticate with your Git Provider{" "}
    <Button variant="contained" color="primary" onClick={onClick} type="button">
      Authenticate with Github
    </Button>
  </Flex>
);

function AuthAlert({ className, onClick, title }: Props) {
  return (
    <Alert
      className={className}
      severity="error"
      title={title}
      message={<Message onClick={onClick} />}
    />
  );
}

export default styled(AuthAlert).attrs({ className: AuthAlert.name })`
  ${Button} {
    margin-left: 16px;
  }
`;
