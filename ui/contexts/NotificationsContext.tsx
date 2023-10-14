import { createContext, Dispatch, useContext } from 'react';

export interface NotificationData {
  message: { text?: string; component?: JSX.Element };
  severity?: 'success' | 'error' | 'warning' | 'info';
  display?: 'top' | 'bottom';
}

 type NotificationContext = {
  notifications: NotificationData[] | [];
  setNotifications: Dispatch<React.SetStateAction<NotificationData[] | []>>;
};

export const Notification = createContext<NotificationContext | null>(null);

export default () => useContext(Notification) as NotificationContext;
