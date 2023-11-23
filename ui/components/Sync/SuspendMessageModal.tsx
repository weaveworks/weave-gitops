import React, { Dispatch, SetStateAction } from "react";
import Modal from "../Modal";
import { TextField } from "@material-ui/core";
import styled from "styled-components";

export type Props = {
  onCloseModal: Dispatch<SetStateAction<boolean>>;
  open: boolean;
  setSuspendMessage: Dispatch<SetStateAction<string>>;
  suspend: any;
  suspendMessage: string;
};

const MessageModal = styled(Modal)`
  &.test {
    background-color: red;
  }
  & .test {
    background-color: red;
  }
  .test {
    background-color: red;
  }
  background-color: red;
`;

function SuspendMessageModal({
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
      className="test"
      bodyClassName="test"
    />
  );
}

export default SuspendMessageModal;
