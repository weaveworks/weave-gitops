import MaterialModal from "@mui/material/Modal";
import * as React from "react";
import styled from "styled-components";
import { IconButton } from "./Button";
import Flex from "./Flex";
import Icon, { IconType } from "./Icon";

/** Modal Properties */
export interface Props {
  /** CSS MUI Overrides or other styling. (for the `<div />` that wraps Modal) */
  className?: string;
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
  background-color: ${(props) => props.theme.colors.neutral00};
  margin: 0 auto;
  max-width: 540px;
  padding: 16px 32px;

  // vertically center using transform:
  transform: translateY(-50%);
  top: 50%;
  position: relative;

  max-height: 90vh;
`;

/** Form Modal */
function UnstyledModal({
  className,
  open,
  onClose,
  title,
  description,
  children,
}: Props) {
  return (
    <MaterialModal
      open={open}
      onClose={onClose}
      aria-labelledby="simple-modal-title"
      aria-describedby="simple-modal-description"
    >
      <Body className={className}>
        <Flex column>
          <Flex row wide align between>
            <h2 id="simple-modal-title">{title}</h2>
            <IconButton
              onClick={onClose}
              className={className}
              variant="text"
              color="inherit"
              size="large"
            >
              <Icon type={IconType.ClearIcon} size="medium" color="neutral30" />
            </IconButton>
          </Flex>

          <p id="simple-modal-description">{description}</p>
        </Flex>
        {children}
      </Body>
    </MaterialModal>
  );
}

export const Modal = styled(UnstyledModal)``;
export default Modal;
