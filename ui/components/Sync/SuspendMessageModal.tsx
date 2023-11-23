import { TextField } from "@material-ui/core";
import React, { Dispatch, SetStateAction } from "react";
import styled from "styled-components";
import Modal from "../Modal";

export type Props = {
  onCloseModal: Dispatch<SetStateAction<boolean>>;
  open: boolean;
  setSuspendMessage: Dispatch<SetStateAction<string>>;
  suspend: any;
  suspendMessage: string;
  className?: string;
};

const MessageModal = styled(Modal)`
  background-color: red;
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
    setSuspendMessage(suspendMessage);
    suspend.mutateAsync();
    setSuspendMessage("");
    onCloseModal(false);
  };

  const onClose = () => closeHandler();

  const content = (
    <form>
      <TextField
        value={suspendMessage}
        onChange={(ev) => setSuspendMessage(ev.target.value)}
      ></TextField>
    </form>
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
