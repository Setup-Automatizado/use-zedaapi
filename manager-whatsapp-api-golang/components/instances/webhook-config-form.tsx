"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Loader2, Save } from "lucide-react";
import { useState, useTransition } from "react";
import { useForm, useWatch } from "react-hook-form";
import { toast } from "sonner";
import { updateWebhookSettings } from "@/actions";
import { Button } from "@/components/ui/button";
import {
	Form,
	FormControl,
	FormDescription,
	FormField,
	FormItem,
	FormLabel,
} from "@/components/ui/form";
import { Separator } from "@/components/ui/separator";
import { Switch } from "@/components/ui/switch";
import {
	type WebhookConfig,
	type WebhookConfigInput,
	WebhookConfigSchema,
} from "@/schemas";
import { WebhookField } from "./webhook-field";

export interface WebhookConfigFormProps {
	instanceId: string;
	instanceToken: string;
	initialValues?: Partial<WebhookConfig>;
}

/**
 * WebhookConfigForm Component
 *
 * Form for configuring all 7 webhook URLs for a WhatsApp instance.
 * Includes individual save buttons and a bulk save option.
 *
 * Features:
 * - Individual webhook URL inputs with validation
 * - notifySentByMe toggle
 * - Save all webhooks at once
 * - Clear all webhooks
 * - Real-time validation
 * - Loading states
 * - Error handling
 *
 * @param instanceId - WhatsApp instance ID
 * @param instanceToken - Instance authentication token
 * @param initialValues - Initial form values
 */
export function WebhookConfigForm({
	instanceId,
	instanceToken,
	initialValues,
}: WebhookConfigFormProps) {
	const [isPending, startTransition] = useTransition();
	const [isClearing, setIsClearing] = useState(false);

	const form = useForm<WebhookConfigInput, unknown, WebhookConfig>({
		resolver: zodResolver(WebhookConfigSchema),
		defaultValues: {
			deliveryCallbackUrl: initialValues?.deliveryCallbackUrl || "",
			receivedCallbackUrl: initialValues?.receivedCallbackUrl || "",
			receivedAndDeliveryCallbackUrl:
				initialValues?.receivedAndDeliveryCallbackUrl || "",
			messageStatusCallbackUrl: initialValues?.messageStatusCallbackUrl || "",
			connectedCallbackUrl: initialValues?.connectedCallbackUrl || "",
			disconnectedCallbackUrl: initialValues?.disconnectedCallbackUrl || "",
			presenceChatCallbackUrl: initialValues?.presenceChatCallbackUrl || "",
			notifySentByMe: initialValues?.notifySentByMe || false,
		},
	});

	const {
		formState: { errors, isDirty },
		setValue,
		reset,
		control,
	} = form;

	const notifySentByMe = useWatch({ control, name: "notifySentByMe" });

	const webhookFields = [
		{
			name: "deliveryCallbackUrl" as const,
			label: "Delivery Webhook",
			description: "Notifications when messages are delivered to the recipient",
		},
		{
			name: "receivedCallbackUrl" as const,
			label: "Received Webhook",
			description: "Notifications when messages are received",
		},
		{
			name: "receivedAndDeliveryCallbackUrl" as const,
			label: "Received and Delivery Webhook",
			description: "Combined notifications for received and delivered messages",
		},
		{
			name: "messageStatusCallbackUrl" as const,
			label: "Message Status Webhook",
			description: "Notifications about message status changes",
		},
		{
			name: "connectedCallbackUrl" as const,
			label: "Connection Webhook",
			description: "Notifications when the instance connects to WhatsApp",
		},
		{
			name: "disconnectedCallbackUrl" as const,
			label: "Disconnection Webhook",
			description: "Notifications when the instance disconnects from WhatsApp",
		},
		{
			name: "presenceChatCallbackUrl" as const,
			label: "Chat Presence Webhook",
			description: "Notifications about presence status (typing, online, etc.)",
		},
	];

	const onSubmit = async (data: WebhookConfig) => {
		startTransition(async () => {
			try {
				const result = await updateWebhookSettings(
					instanceId,
					instanceToken,
					data,
				);

				if (result.success) {
					toast.success("Webhooks updated successfully");
					reset(data); // Reset form with new values to clear dirty state
				} else {
					toast.error(result.error || "Error updating webhooks");
				}
			} catch (error) {
				console.error("Error updating webhooks:", error);
				toast.error("Error updating webhooks");
			}
		});
	};

	/**
	 * Check if any webhook URL is configured (not empty)
	 * Used to enable/disable Clear All button independently of dirty state
	 */
	const hasConfiguredWebhooks = webhookFields.some(
		(field) => form.getValues(field.name),
	);

	/**
	 * Clear all webhook URLs and persist to server
	 * Unlike just resetting the form, this actually calls the API to clear webhooks
	 */
	const handleClearAll = async () => {
		setIsClearing(true);

		// Clear all webhook URLs but keep notifySentByMe
		const clearedValues: WebhookConfigInput = {
			deliveryCallbackUrl: "",
			receivedCallbackUrl: "",
			receivedAndDeliveryCallbackUrl: "",
			messageStatusCallbackUrl: "",
			connectedCallbackUrl: "",
			disconnectedCallbackUrl: "",
			presenceChatCallbackUrl: "",
			notifySentByMe: notifySentByMe ?? false,
		};

		try {
			// Call API to persist the clearing
			const result = await updateWebhookSettings(
				instanceId,
				instanceToken,
				clearedValues,
			);

			if (result.success) {
				reset(clearedValues);
				toast.success("All webhooks cleared successfully");
			} else {
				toast.error(result.error || "Failed to clear webhooks");
			}
		} catch (error) {
			console.error("Error clearing webhooks:", error);
			toast.error("Error clearing webhooks");
		} finally {
			setIsClearing(false);
		}
	};

	return (
		<Form {...form}>
			<form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
				{/* Webhook URL Fields */}
				<div className="space-y-6">
					{webhookFields.map((field) => (
						<FormField
							key={field.name}
							control={form.control}
							name={field.name}
							render={({ field: formField }) => (
								<FormItem>
									<WebhookField
										name={field.name}
										label={field.label}
										description={field.description}
										value={formField.value || ""}
										error={errors[field.name]?.message}
										onChange={formField.onChange}
										onClear={() =>
											setValue(field.name, "", {
												shouldDirty: true,
											})
										}
										disabled={isPending || isClearing}
									/>
								</FormItem>
							)}
						/>
					))}
				</div>

				<Separator />

				{/* Notify Sent By Me Toggle */}
				<FormField
					control={form.control}
					name="notifySentByMe"
					render={({ field }) => (
						<FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
							<div className="space-y-0.5">
								<FormLabel className="text-base">
									Notify Sent Messages
								</FormLabel>
								<FormDescription>
									Receive webhooks for messages sent by this instance
								</FormDescription>
							</div>
							<FormControl>
								<Switch
									checked={field.value}
									onCheckedChange={field.onChange}
									disabled={isPending || isClearing}
								/>
							</FormControl>
						</FormItem>
					)}
				/>

				{/* Action Buttons */}
				<div className="flex flex-col-reverse gap-3 sm:flex-row sm:justify-between">
					<Button
						type="button"
						variant="outline"
						onClick={handleClearAll}
						disabled={isPending || isClearing || !hasConfiguredWebhooks}
					>
						{isClearing ? (
							<>
								<Loader2 className="mr-2 h-4 w-4 animate-spin" />
								Clearing...
							</>
						) : (
							"Clear All"
						)}
					</Button>

					<div className="flex gap-3">
						<Button
							type="button"
							variant="outline"
							onClick={() => reset()}
							disabled={isPending || isClearing || !isDirty}
						>
							Cancel
						</Button>
						<Button
							type="submit"
							disabled={isPending || isClearing || !isDirty}
						>
							{isPending ? (
								<>
									<Loader2 className="mr-2 h-4 w-4 animate-spin" />
									Saving...
								</>
							) : (
								<>
									<Save className="mr-2 h-4 w-4" />
									Save All
								</>
							)}
						</Button>
					</div>
				</div>

				{/* Dirty State Indicator */}
				{isDirty && !isPending && (
					<p className="text-sm text-muted-foreground text-center">
						You have unsaved changes
					</p>
				)}
			</form>
		</Form>
	);
}
