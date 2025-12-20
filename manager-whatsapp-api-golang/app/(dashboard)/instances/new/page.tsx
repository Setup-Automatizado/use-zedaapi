import type { Metadata } from "next";
import { CreateInstanceForm } from "@/components/instances/create-instance-form";
import { PageHeader } from "@/components/shared/page-header";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";

export const metadata: Metadata = {
	title: "New Instance | WhatsApp Manager",
	description: "Create new WhatsApp instance",
};

export default function NewInstancePage() {
	return (
		<div className="space-y-6">
			<PageHeader
				title="New Instance"
				description="Configure a new WhatsApp instance to start sending and receiving messages"
			/>

			<Card>
				<CardHeader>
					<CardTitle>Instance Configuration</CardTitle>
					<CardDescription>
						Fill in the basic information and configure optional webhooks and
						preferences
					</CardDescription>
				</CardHeader>
				<CardContent>
					<CreateInstanceForm />
				</CardContent>
			</Card>
		</div>
	);
}
