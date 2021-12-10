import MaterialModal from "@material-ui/core/Modal";
import * as React from "react";
import styled from "styled-components";
import Button from "./Button";
import Flex from "./Flex";

/** Modal Properties */
export interface Props {
  /** CSS MUI Overrides or other styling. (for the `<div />` that wraps Modal) */
  className?: string;
  /** CSS MUI Overrides or other styling. (for the Modal `<Body />`) */
  bodyClassName?: string;
  /** state variable to display Modal */
  open: boolean;
  /** Close event handler function */
  onClose: () => void;
  /** `<h2 />` element appearing at the top of the modal */
  title: string;
  /** `<p />` element appearing below title */
  description: string;
  /** Modal content */
  children: any;
}

export const Body = styled.div`
  display: flex;
  flex-direction: column;
  justify-content: space-evenly;
  background-color: ${(props) => props.theme.colors.white};
  margin: 0 auto;
  max-width: 540px;
  padding: 16px 32px;
  transform: translate(0, 50%);
`;

/** Form Modal */
function UnstyledModal({
  className,
  bodyClassName,
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
        <Body className={bodyClassName}>
          <Flex column>
            <h2 id="simple-modal-title">{title}</h2>
            <p id="simple-modal-description">{description}</p>
          </Flex>
          <div>{children}</div>
          <Flex wide end>
            <Button onClick={onClose} color="inherit" variant="text">
              Close
            </Button>
          </Flex>
        </Body>
      </MaterialModal>
    </div>
  );
}

export const Modal = styled(UnstyledModal)``;
export default Modal;
