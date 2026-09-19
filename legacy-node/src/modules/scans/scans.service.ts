/**
 * Scan/upload orchestration.
 *
 * The API path here does the CHEAP work only: store the image in object storage,
 * create the authoritative DB rows, and enqueue a BullMQ job. All expensive
 * OCR/ML work happens later in the workers. The API returns a job id immediately.
 */
import { prisma } from '../../config/database';
import { NotFoundError } from '../../common/errors';
import { buildObjectKey, putObject } from '../../config/storage';
import { sha256 } from '../../common/utils/crypto';
import { scanQueue, QUEUE_NAMES } from '../../config/queues';
import { ensureInspection } from '../inspections/inspections.service';

export interface UploadedFile {
  buffer: Buffer;
  originalname: string;
  mimetype: string;
  size: number;
}

/** Store an uploaded image and enqueue it for processing. */
export async function addImage(inspectionId: string, file: UploadedFile) {
  await ensureInspection(inspectionId);

  const key = buildObjectKey(`inspections/${inspectionId}/original`, file.originalname);
  await putObject({ key, body: file.buffer, contentType: file.mimetype });

  const image = await prisma.inspectionImage.create({
    data: {
      inspectionId,
      storageKey: key,
      originalFilename: file.originalname,
      mimeType: file.mimetype,
      sizeBytes: file.size,
      kind: 'ORIGINAL',
      status: 'UPLOADED',
      checksum: sha256(file.buffer),
    },
  });

  const job = await enqueueImage(inspectionId, image.id);

  await prisma.inspectionImage.update({ where: { id: image.id }, data: { status: 'QUEUED' } });

  return { image, jobId: job.mlJobId, bullJobId: job.bullJobId };
}

/** (Re)process every image on an inspection. */
export async function processInspection(inspectionId: string) {
  await ensureInspection(inspectionId);

  const images = await prisma.inspectionImage.findMany({
    where: { inspectionId, kind: 'ORIGINAL' },
    select: { id: true },
  });

  const jobs = [];
  for (const img of images) {
    jobs.push(await enqueueImage(inspectionId, img.id));
  }

  await prisma.inspection.update({
    where: { id: inspectionId },
    data: { status: 'PROCESSING', startedAt: new Date() },
  });

  return { queued: jobs.length, jobs };
}

/** List the ML jobs (processing state) for an inspection. */
export async function listJobs(inspectionId: string) {
  await ensureInspection(inspectionId);
  return prisma.mLJob.findMany({
    where: { inspectionId },
    orderBy: { createdAt: 'desc' },
  });
}

/** Create an MLJob row + BullMQ job for a single image. */
async function enqueueImage(inspectionId: string, imageId: string) {
  const mlJob = await prisma.mLJob.create({
    data: {
      inspectionId,
      imageId,
      type: 'ANALYZE',
      status: 'QUEUED',
      queueName: QUEUE_NAMES.SCAN,
    },
  });

  const bull = await scanQueue.add(
    'process-image',
    { inspectionId, imageId, mlJobId: mlJob.id },
    { jobId: mlJob.id },
  );

  await prisma.mLJob.update({ where: { id: mlJob.id }, data: { bullJobId: bull.id } });

  return { mlJobId: mlJob.id, bullJobId: bull.id };
}
