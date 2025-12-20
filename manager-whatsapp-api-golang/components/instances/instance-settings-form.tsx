"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Loader2, Save } from "lucide-react";
import { useTransition } from "react";
import { useForm, useWatch } from "react-hook-form";
import { toast } from "sonner";
import { updateInstanceSettings } from "@/actions";
import { Button } from "@/components/ui/button";
import {
	Form,
	FormControl,
	FormDescription,
	FormField,
	FormItem,
	FormLabel,
	FormMessage,
} from "@/components/ui/form";
import { Separator } from "@/components/ui/separator";
import { Switch } from "@/components/ui/switch";
import { Textarea } from "@/components/ui/textarea";
import { cn } from "@/lib/utils";
import {
	type InstanceSettings,
	type InstanceSettingsInput,
	InstanceSettingsSchema,
} from "@/schemas";

export interface InstanceSettingsFormProps {
	instanceId: string;
	instanceToken: string;
	initialValues?: Partial<InstanceSettings>;
}

/**
 * InstanceSettingsForm Component
 *
 * Form for configuring instance behavior settings including:
 * - Auto read messages
 * - Call rejection (auto and custom message)
 *
 * Features:
 * - Real-time validation
 * - Conditional fields (reject message shown when auto reject is enabled)
 * - Loading states
 * - Error handling
 * - Optimistic updates
 *
 * @param instanceId - WhatsApp instance ID
 * @param instanceToken - Instance authentication token
 * @param initialValues - Initial form values
 */
export function InstanceSettingsForm({
	instanceId,
	instanceToken,
	initialValues,
}: InstanceSettingsFormProps) {
	const [isPending, startTransition] = useTransition();

	const form = useForm<InstanceSettingsInput, unknown, InstanceSettings>({
		resolver: zodResolver(InstanceSettingsSchema),
		defaultValues: {
			autoReadMessage: initialValues?.autoReadMessage || false,
			callRejectAuto: initialValues?.callRejectAuto || false,
			callRejectMessage: initialValues?.callRejectMessage || "",
			notifySentByMe: initialValues?.notifySentByMe || false,
		},
	});

	const {
		formState: { errors, isDirty },
		reset,
		control,
	} = form;

	const callRejectAuto = useWatch({ control, name: "callRejectAuto" });

	const onSubmit = async (data: InstanceSettings) => {
		startTransition(async () => {
			try {
				const result = await updateInstanceSettings(
					instanceId,
					instanceToken,
					data,
				);

				if (result.success) {
					toast.success("Settings updated successfully");
					reset(data); // Reset form with new values to clear dirty state
				} else {
					toast.error(result.error || "Error updating settings");
				}
			} catch (error) {
				console.error("Error updating settings:", error);
				toast.error("Error updating settings");
			}
		});
	};

	return (
		<Form {...form}>
			<form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
				{/* Auto Read Message */}
				<FormField
					control={form.control}
					name="autoReadMessage"
					render={({ field }) => (
						<FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
							<div className="space-y-0.5">
								<FormLabel className="text-base">
									Mark Messages as Read
								</FormLabel>
								<FormDescription>
									Automatically mark all received messages as read
								</FormDescription>
							</div>
							<FormControl>
								<Switch
									checked={field.value}
									onCheckedChange={field.onChange}
									disabled={isPending}
								/>
							</FormControl>
						</FormItem>
					)}
				/>

				<Separator />

				{/* Call Reject Auto */}
				<FormField
					control={form.control}
					name="callRejectAuto"
					render={({ field }) => (
						<FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
							<div className="space-y-0.5">
								<FormLabel className="text-base">
									Reject Calls Automatically
								</FormLabel>
								<FormDescription>
									Automatically reject all incoming calls
								</FormDescription>
							</div>
							<FormControl>
								<Switch
									checked={field.value}
									onCheckedChange={field.onChange}
									disabled={isPending}
								/>
							</FormControl>
						</FormItem>
					)}
				/>

				{/* Call Reject Message - Shown only when callRejectAuto is enabled */}
				{callRejectAuto && (
					<FormField
						control={form.control}
						name="callRejectMessage"
						render={({ field }) => (
							<FormItem>
								<FormLabel>
									Call Rejection Message
									<span className="text-destructive ml-1">*</span>
								</FormLabel>
								<FormDescription>
									Message sent automatically when a call is rejected. Maximum
									500 characters.
								</FormDescription>
								<FormControl>
									<Textarea
										placeholder="Sorry, I can't answer calls right now. Please send a text message."
										className={cn(
											"resize-none",
											errors.callRejectMessage && "border-destructive",
										)}
										rows={4}
										maxLength={500}
										disabled={isPending}
										{...field}
										value={field.value || ""}
									/>
								</FormControl>
								<div className="flex items-center justify-between">
									<FormMessage />
									<p className="text-sm text-muted-foreground">
										{field.value?.length || 0}/500
									</p>
								</div>
							</FormItem>
						)}
					/>
				)}

				<Separator />

				{/* Notify Sent By Me */}
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
									Receive notifications for messages sent by this instance
								</FormDescription>
							</div>
							<FormControl>
								<Switch
									checked={field.value}
									onCheckedChange={field.onChange}
									disabled={isPending}
								/>
							</FormControl>
						</FormItem>
					)}
				/>

				{/* Action Buttons */}
				<div className="flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
					<Button
						type="button"
						variant="outline"
						onClick={() => reset()}
						disabled={isPending || !isDirty}
					>
						Cancel
					</Button>
					<Button type="submit" disabled={isPending || !isDirty}>
						{isPending ? (
							<>
								<Loader2 className="mr-2 h-4 w-4 animate-spin" />
								Saving...
							</>
						) : (
							<>
								<Save className="mr-2 h-4 w-4" />
								Save Settings
							</>
						)}
					</Button>
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
