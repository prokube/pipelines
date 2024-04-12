import argparse
import subprocess


def get_current_branch_name():
    try:
        result = subprocess.check_output(
            ["git", "rev-parse", "--abbrev-ref", "HEAD"], stderr=subprocess.PIPE
        )
        branch_name = result.decode("utf-8").strip()
        return branch_name
    except subprocess.CalledProcessError:
        return None


def main(commit_message: str, images: str) -> None:
    # images should be in square brackets in accordance with workflow condition
    commit_command = ["git", "commit", "-m", f"{commit_message} [{images}]"]
    subprocess.run(commit_command, check=True)

    current_branch = get_current_branch_name()
    if current_branch is None:
        raise Exception(
            "Cannot push to upstream branch. "
            + "Please make sure you are checked out to some branch. "
            + "Otherwise run the push command manually."
        )

    push_command = ["git", "push", "origin", current_branch]
    subprocess.run(push_command, check=True)


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "--images",
        choices=["backend-all", "apiserver", "driver-launcher"],
        required=True,
    )
    parser.add_argument("--commit-message", default="Trigger image build:")
    args = parser.parse_args()
    main(commit_message=args.commit_message, images=args.images)
