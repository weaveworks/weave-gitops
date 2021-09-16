import { Button } from "@material-ui/core";
import MaterialModal from "@material-ui/core/Modal";
import * as React from "react";
import styled from "styled-components";
import Flex from "./Flex";

type Props = {
  className?: string;
  open: boolean;
  onClose: () => void;
  title: string;
  description: string;
  children: any;
};

const Body = styled.div`
  background-color: ${(props) => props.theme.colors.white};
  margin: 0 auto;
  max-width: 540px;
  padding: 16px 32px;
  transform: translate(0, 50%);
`;

function Modal({
  className,
  open,
  onClose,
  title,
  description,
  children,
}: Props) {
  return (
    <div className={className}>
      <MaterialModal
        open={open}
        onClose={onClose}
        aria-labelledby="simple-modal-title"
        aria-describedby="simple-modal-description"
      >
        <Body>
          <h2 id="simple-modal-title">{title}</h2>
          <p id="simple-modal-description">{description}</p>
          <div>{children}</div>
          <Flex wide end>
            <Button variant="contained" onClick={onClose}>
              Close
            </Button>
          </Flex>
        </Body>
      </MaterialModal>
    </div>
  );
}

export default styled(Modal)``;
