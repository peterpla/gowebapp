#!/usr/bin/bash

# un-comment commands you need to perform, comment those already done
# see "Cloud Task queues with App Engine targets"
# <https://cloud.google.com/tasks/docs/dual-overview#appe>

# List all available queues
gcloud tasks queues list

# TODO: if desired queue is listed, it already exists; exit

# 1. "Create a worker to process the tasks" - DONE elsewhere
# Create an App Engine service using app.yaml's "service:" element.
# (Must be separate from the "default" service)

# 2. "Create a queue"
# <https://cloud.google.com/tasks/docs/creating-queues>
#gcloud tasks queues create wInitialRequest

# "It can take a few minutes for a newly created queue to be available."
# TODO: sleep or loop until queue listed?

# "use describe to verify that your queue was created successfully"
#gcloud tasks queues describe wInitialRequest

# NOTE: The remaining steps described are implemented by the application.