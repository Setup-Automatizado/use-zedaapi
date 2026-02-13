"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Loader2 } from "lucide-react";
import { useTransition } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
import {
	Form,
	FormControl,
	FormField,
	FormItem,
	FormLabel,
	FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import {
	CreateGroupSchema,
	type CreateGroupFormValues,
	type CreateGroupInput,
} from "@/schemas/pool";

interface GroupFormProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	onSubmit: (data: CreateGroupFormValues) => Promise<void>;
}

export function GroupForm({ open, onOpenChange, onSubmit }: GroupFormProps) {
	const [isPending, startTransition] = useTransition();

	const form = useForm<CreateGroupInput, unknown, CreateGroupFormValues>({
		resolver: zodResolver(CreateGroupSchema),
		defaultValues: {
			name: "",
			maxInstances: 10,
		},
	});

	const handleSubmit = (values: CreateGroupFormValues) => {
		startTransition(async () => {
			try {
				await onSubmit(values);
				form.reset();
				onOpenChange(false);
				toast.success("Group created");
			} catch {
				toast.error("Failed to create group");
			}
		});
	};

	return (
		<Dialog open={open} onOpenChange={onOpenChange}>
			<DialogContent className="sm:max-w-md">
				<DialogHeader>
					<DialogTitle>Create Proxy Group</DialogTitle>
					<DialogDescription>
						Groups allow multiple instances to share a single proxy.
					</DialogDescription>
				</DialogHeader>

				<Form {...form}>
					<form
						onSubmit={form.handleSubmit(handleSubmit)}
						className="space-y-4"
					>
						<FormField
							control={form.control}
							name="name"
							render={({ field }) => (
								<FormItem>
									<FormLabel>Group Name</FormLabel>
									<FormControl>
										<Input
											placeholder="Brazil Group"
											{...field}
										/>
									</FormControl>
									<FormMessage />
								</FormItem>
							)}
						/>

						<FormField
							control={form.control}
							name="maxInstances"
							render={({ field }) => (
								<FormItem>
									<FormLabel>Max Instances</FormLabel>
									<FormControl>
										<Input
											type="number"
											{...field}
											onChange={(e) =>
												field.onChange(
													Number(e.target.value),
												)
											}
										/>
									</FormControl>
									<FormMessage />
								</FormItem>
							)}
						/>

						<FormField
							control={form.control}
							name="countryCode"
							render={({ field }) => (
								<FormItem>
									<FormLabel>
										Country Code (optional)
									</FormLabel>
									<FormControl>
										<Input
											placeholder="BR"
											maxLength={2}
											{...field}
										/>
									</FormControl>
									<FormMessage />
								</FormItem>
							)}
						/>

						<DialogFooter>
							<Button
								type="button"
								variant="outline"
								onClick={() => onOpenChange(false)}
							>
								Cancel
							</Button>
							<Button type="submit" disabled={isPending}>
								{isPending && (
									<Loader2 className="mr-2 h-4 w-4 animate-spin" />
								)}
								Create Group
							</Button>
						</DialogFooter>
					</form>
				</Form>
			</DialogContent>
		</Dialog>
	);
}
