import {PageShell} from "../../../../components/PageShell";
import {UserForm} from "../../../../features/users/UserForm";

export default function NewUserPage() {
  return <PageShell title="Create user"><UserForm/></PageShell>;
}
