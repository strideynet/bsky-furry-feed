const SHOW_QUEUE_ACTION_CONFIRMATION = "bff-show-queue-action-confirmation";

export const showQueueActionConfirmation = ref(
  localStorage.getItem(SHOW_QUEUE_ACTION_CONFIRMATION) === "true"
);

watch(showQueueActionConfirmation, (show) => {
  if (show) {
    localStorage.setItem(SHOW_QUEUE_ACTION_CONFIRMATION, "true");
  } else {
    localStorage.removeItem(SHOW_QUEUE_ACTION_CONFIRMATION);
  }
});
