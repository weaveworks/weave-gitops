import { Meta } from "@storybook/react";
import React from "react";
import Button from "../components/Button";
import Modal from "../components/Modal";

export default {
  title: "Modal",
  component: Modal,
  argTypes: {
    title: {
      description:
        //PUT IN LINK TO MUI MODAL
        "`<h2 />` element appearing at the top of the modal",
      defaultValue: "Title",
      type: { summary: "string", name: "string", required: false },
    },
    description: {
      description: "`<p />` element appearing below title",
      defaultValue: "Description",
      type: { summary: "string", name: "string", required: false },
    },
    className: {
      description: "CSS MUI Overrides or other styling",
      type: { summary: "string", name: "string", required: false },
    },
    bodyClassName: {
      description: "CSS MUI Overrides or other styling for Modal children",
      type: { summary: "string", name: "string", required: false },
    },
    children: {
      description: "React Nodes to make up Modal body",
    },
  },
} as Meta;

export const Default: React.VFC<{
  title: "Title";
  description: "description";
}> = (args) => {
  const [open, setOpen] = React.useState(false);
  return (
    <div>
      {open ? (
        <Modal open={open} onClose={() => setOpen(false)} {...args}>
          <div>This is the Modal body!</div>
        </Modal>
      ) : (
        <Button onClick={() => setOpen(true)}>Open Modal</Button>
      )}
    </div>
  );
};
