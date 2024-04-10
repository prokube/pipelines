import argparse
import subprocess


def main(commit_message: str, images: str) -> None:
    # images should be in square brackets in accordance with workflow condition
    commit_command = ['git', 'commit', '-m', f"{commit_message} [{images}]"]
    subprocess.run(commit_command, check=True)

    push_command = ['git', 'push', 'origin', 'GCP-artifact-registry']
    subprocess.run(push_command, check=True)


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--images", choices=["backend-all", "apiserver"], required=True)
    parser.add_argument("--commit-message", default="Trigger image build:")
    args = parser.parse_args()
    main(commit_message=args.commit_message, images=args.images)
