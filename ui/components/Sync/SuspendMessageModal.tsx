import React, { Dispatch, SetStateAction } from "react";
import Modal from "../Modal";
import Input from "../Input";

export type Props = {
  onCloseModal: Dispatch<SetStateAction<boolean>>;
  open: boolean;
};
let suspendMessage ='';


function SuspendMessageModal({ onCloseModal, open,  }: Props) {
  const onClose = () => onCloseModal(false);

  const content = (
    <Input value={suspendMessage}></Input>
  );

  return (
    <Modal
      open={open}
      onClose={onClose}
      title="Suspend message"
      description="Add a suspend message"
      children={content}
    />
  );
}

export default SuspendMessageModal;
