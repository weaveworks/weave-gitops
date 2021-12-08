import { Meta, Story } from "@storybook/react";
import React from "react";
import Button from "../components/Button";
import Modal, { Props } from "../components/Modal";

export default {
  title: "Modal",
  component: Modal,
  parameters: {
    docs: {
      description: {
        component: "MUI Modal - see https://mui.com/components/modal/",
      },
    },
  },
} as Meta;

const Template: Story<Props> = (args) => {
  const [open, setOpen] = React.useState(false);
  return (
    <div>
      {open ? (
        <Modal {...args} open={open} onClose={() => setOpen(false)}>
          <div>This is the Modal body!</div>
        </Modal>
      ) : (
        <Button onClick={() => setOpen(true)}>Open Modal</Button>
      )}
    </div>
  );
};
export const Default = Template.bind({});
Default.args = { title: "Title", description: "Description" };
