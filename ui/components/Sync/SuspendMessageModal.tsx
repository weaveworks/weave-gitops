import React, { Dispatch, SetStateAction } from "react";
import styled from "styled-components";
import Modal from "../Modal";
import Flex from "../Flex";
import Button from "../Button";

export type Props = {
  onCloseModal: Dispatch<SetStateAction<boolean>>;
  open: boolean;
  setSuspendMessage: Dispatch<SetStateAction<string>>;
  suspend: any;
  suspendMessage: string;
  className?: string;
};

const MessageModal = styled(Modal)`
  textarea {
    width: 100%;
    box-sizing: border-box;
    font-family: inherit;
    font-size: 100%;
    border-radius: ${(props) => props.theme.spacing.xxs};
    resize: none;
    margin-bottom: ${(props) => props.theme.spacing.base};
  }
`;

function SuspendMessageModal({
  className,
  onCloseModal,
  open,
  setSuspendMessage,
  suspend,
  suspendMessage,
}: Props) {
  const closeHandler = () => {
    onCloseModal(false);
  };
  const suspendHandler = () => {
    setSuspendMessage(suspendMessage);
    suspend.mutateAsync();
    setSuspendMessage("");
    onCloseModal(false);
  };

  const onClose = () => closeHandler();

  const content = (
    <>
      <textarea
        rows={5}
        value={suspendMessage}
        onChange={(ev) => setSuspendMessage(ev.target.value)}
      ></textarea>
      <Flex wide end>
        <Button onClick={suspendHandler} color="inherit" variant="text">
          Suspend
        </Button>
      </Flex>
    </>
  );

  return (
    <MessageModal
      open={open}
      onClose={onClose}
      title="Suspend Message"
      description="Add reaasdasdson for suspending"
      children={content}
      className={className}
    />
  );
}

export default SuspendMessageModal;
